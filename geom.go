package pgxorb

import (
	"context"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/ewkb"
)

var orgGeometryInterfaceType = reflect.TypeOf((*orb.Geometry)(nil)).Elem()

type geometryCodec struct{}

// A geometryBinaryEncodePlan implements
// [github.com/jackc/pgx/v5/pgtype.EncodePlan] for types in binary format.
type geometryBinaryEncodePlan struct{}

// A geometryTextEncodePlan implements
// [github.com/jackc/pgx/v5/pgtype.EncodePlan] for types in text format.
type geometryTextEncodePlan struct{}

// A geometryBinaryScanPlan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan]
type geometryBinaryScanPlan struct{}

// A geometryTextScanPlan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan]
type geometryTextScanPlan struct{}

// FormatSupported implements
// [github.com/jackc/pgx/v5/pgtype.Codec.FormatSupported].
func (c *geometryCodec) FormatSupported(format int16) bool {
	switch format {
	case pgtype.BinaryFormatCode:
		return true
	case pgtype.TextFormatCode:
		return true
	default:
		return false
	}
}

// PreferredFormat implements
// [github.com/jackc/pgx/v5/pgtype.Codec.PreferredFormat].
func (c *geometryCodec) PreferredFormat() int16 {
	return pgtype.BinaryFormatCode
}

// PlanEncode implements [github.com/jackc/pgx/v5/pgtype.Codec.PlanEncode].
func (c *geometryCodec) PlanEncode(m *pgtype.Map, old uint32, format int16, value any) pgtype.EncodePlan {
	switch format {
	case pgtype.BinaryFormatCode:
		return geometryBinaryEncodePlan{}
	case pgtype.TextFormatCode:
		return geometryTextEncodePlan{}
	default:
		return nil
	}
}

// PlanScan implements [github.com/jackc/pgx/v5/pgtype.Codec.PlanScan].
func (c *geometryCodec) PlanScan(m *pgtype.Map, old uint32, format int16, target any) pgtype.ScanPlan {
	switch format {
	case pgx.BinaryFormatCode:
		return geometryBinaryScanPlan{}
	case pgx.TextFormatCode:
		return geometryTextScanPlan{}
	default:
		return nil
	}
}

// DecodeDatabaseSQLValue implements
// [github.com/jackc/pgx/v5/pgtype.Codec.DecodeDatabaseSQLValue].
func (c *geometryCodec) DecodeDatabaseSQLValue(
	m *pgtype.Map,
	oid uint32,
	format int16,
	src []byte,
) (driver.Value, error) {
	return nil, errors.ErrUnsupported
}

// DecodeValue implements [github.com/jackc/pgx/v5/pgtype.Codec.DecodeValue].
func (c *geometryCodec) DecodeValue(m *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
	switch format {
	case pgtype.TextFormatCode:
		var err error
		src, err = hex.DecodeString(string(src))
		if err != nil {
			return nil, err
		}
		fallthrough
	case pgtype.BinaryFormatCode:
		geom, _, err := ewkb.Unmarshal(src)
		return geom, err
	default:
		return nil, errors.ErrUnsupported
	}
}

// Encode implements [github.com/jackc/pgx/v5/pgtype.EncodePlan.Encode].
func (p geometryBinaryEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	geom, ok := value.(orb.Geometry)
	if !ok {
		return buf, errors.ErrUnsupported
	}

	ewkbBuf, err := ewkb.Marshal(geom, ewkb.DefaultSRID, ewkb.DefaultByteOrder)
	if err != nil {
		return buf, fmt.Errorf("failed to encode geometry: %w", err)
	}

	return append(buf, ewkbBuf...), nil
}

// Encode implements [github.com/jackc/pgx/v5/pgtype.EncodePlan.Encode].
func (p geometryTextEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	geom, ok := value.(orb.Geometry)
	if !ok {
		return buf, errors.ErrUnsupported
	}

	ewkbBuf, err := ewkb.Marshal(geom, ewkb.DefaultSRID, ewkb.DefaultByteOrder)
	if err != nil {
		return buf, fmt.Errorf("failed to encode geometry: %w", err)
	}

	return append(buf, []byte(hex.EncodeToString(ewkbBuf))...), nil
}

// Scan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan.Scan].
func (p geometryBinaryScanPlan) Scan(src []byte, target any) error {
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to a orb.Geometry")
	}

	targetElem := targetType.Elem()
	if !targetElem.Implements(orgGeometryInterfaceType) {
		return fmt.Errorf("target must be a pointer to a orb.Geometry")
	}

	if len(src) == 0 {
		return nil
	}

	geom, _, err := ewkb.Unmarshal(src)
	if err != nil {
		return err
	}

	geomType := reflect.TypeOf(geom)

	if targetType.Elem() != geomType {
		return fmt.Errorf("target type %v doesn't match geometry type %v", targetType, geomType)
	}

	reflect.ValueOf(target).Elem().Set(reflect.ValueOf(geom))

	return nil
}

// Scan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan.Scan].
func (p geometryTextScanPlan) Scan(src []byte, target any) error {
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to a orb.Geometry")
	}

	targetElem := targetType.Elem()
	if !targetElem.Implements(orgGeometryInterfaceType) {
		return fmt.Errorf("target must be a pointer to a orb.Geometry")
	}

	if len(src) == 0 {
		return nil
	}

	var err error
	src, err = hex.DecodeString(string(src))
	if err != nil {
		return err
	}

	geom, _, err := ewkb.Unmarshal(src)
	if err != nil {
		return err
	}

	geomType := reflect.TypeOf(geom)

	if targetType.Elem() != geomType {
		return fmt.Errorf("target type %v doesn't match geometry type %v", targetType, geomType)
	}

	reflect.ValueOf(target).Elem().Set(reflect.ValueOf(geom))

	return nil
}

func registerGeom(ctx context.Context, conn *pgx.Conn) error {
	var geomtypeOID uint32
	err := conn.QueryRow(ctx, "select 'geometry'::text::regtype::oid").Scan(&geomtypeOID)
	if err != nil {
		return fmt.Errorf("get geometry oid failed: %w", err)
	}

	conn.TypeMap().RegisterType(&pgtype.Type{
		Name:  "geometry",
		Codec: &geometryCodec{},
		OID:   geomtypeOID,
	})

	return nil
}
