package dgraph

import (
	"fmt"
	"io"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/dgraph-io/dgraph/types"
	"github.com/twpayne/go-geom"
)

func MarshalDateTime(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		s, _ := t.MarshalJSON()
		w.Write(s)
	})
}

func UnmarshalDateTime(v interface{}) (time.Time, error) {
	t, err := types.ParseTime(v.(string))
	if err != nil {
		fmt.Println("Error parsing time")
		return t, err
	}

	return t, nil
}

func MarshalPoint(p geom.Point) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		w.Write([]byte("point"))
	})
}

func UnmarshalPoint(v interface{}) (geom.Point, error) {
	fmt.Println(v)

	return geom.Point{}, nil
}
