package pgsql

import (
	"github.com/specterops/bloodhound/slicesext"
)

// MustSliceTypeConvert panics on a type conversion error from a slice of type []F as FS to a slice of type []T as TS.
func MustSliceTypeConvert[T any, TS []T, F any, FS []F](fs FS) TS {
	if ts, err := slicesext.MapWithErr(fs, slicesext.ConvertType[F, T]()); err != nil {
		panic(err.Error())
	} else {
		return ts
	}
}
