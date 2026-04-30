package models

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

type Point struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

type ewkbPoint struct {
	ByteOrder byte   // 1 (LittleEndian)
	WkbType   uint32 // 0x20000001 (PointS)
	SRID      uint32 // 4326
	Point     Point
}

func (p *Point) String() string {
	return fmt.Sprintf("SRID=4326;POINT(%v %v)", p.Lng, p.Lat)
}

func (p *Point) Scan(val interface{}) error {
	b, err := hex.DecodeString(val.(string))
	if err != nil {
		return err
	}

	r := bytes.NewReader(b)

	var ewkbP ewkbPoint
	err = binary.Read(r, binary.LittleEndian, &ewkbP)
	if err != nil {
		return err
	}

	if ewkbP.ByteOrder != 1 || ewkbP.WkbType != 0x20000001 || ewkbP.SRID != 4326 {
		return fmt.Errorf("Point.Scan: unexpected ewkb %#v", ewkbP)
	}
	*p = ewkbP.Point
	return nil
}

func (p Point) Value() (driver.Value, error) {
	return p.String(), nil
}

type Polygon []Point

type ewkbPolygon struct {
	ByteOrder byte   // 1 (LittleEndian)
	WkbType   uint32 // 0x20000003 (PolygonS)
	SRID      uint32 // 4326
	Rings     uint32
	Count     uint32
}

func (p *Polygon) String() string {
	points := []string{}
	for _, point := range *p {
		points = append(points, fmt.Sprintf("%v %v", point.Lng, point.Lat))
	}
	points = append(points, fmt.Sprintf("%v %v", (*p)[0].Lng, (*p)[0].Lat))
	return fmt.Sprintf("SRID=4326;POLYGON((%s))", strings.Join(points, ","))
}

func (p *Polygon) Scan(val interface{}) error {
	b, err := hex.DecodeString(val.(string))
	if err != nil {
		return err
	}

	r := bytes.NewReader(b)

	var ewkbP ewkbPolygon
	if err := binary.Read(r, binary.LittleEndian, &ewkbP); err != nil {
		return err
	}

	if ewkbP.ByteOrder != 1 || ewkbP.WkbType != 0x20000003 || ewkbP.SRID != 4326 || ewkbP.Rings != 1 {
		return fmt.Errorf("Polygon.Scan: unexpected ewkb %#v", ewkbP)
	}

	points := make([]Point, ewkbP.Count)
	if err := binary.Read(r, binary.LittleEndian, &points); err != nil {
		return err
	}
	*p = points

	return nil
}

func (p Polygon) Value() (driver.Value, error) {
	return p.String(), nil
}
