package coredbutils

import "github.com/alexbrainman/odbc/api"

// Drv ...
var Drv Driver

// Driver struct to manange env handle
type Driver struct {
	Stats
	H api.SQLHENV
}

// Load the index.html template.
func initDriver() error {
	var out api.SQLHANDLE
	in := api.SQLHANDLE(api.SQL_NULL_HANDLE)
	ret := api.SQLAllocHandle(api.SQL_HANDLE_ENV, in, &out)
	if IsError(ret) {
		return NewError("SQLAllocHandle", api.SQLHENV(in))
	}
	Drv.H = api.SQLHENV(out)
	err := Drv.Stats.UpdateHandleCount(api.SQL_HANDLE_ENV, 1)
	if err != nil {
		return err
	}
	// will use ODBC v3
	ret = api.SQLSetEnvUIntPtrAttr(Drv.H, api.SQL_ATTR_ODBC_VERSION, api.SQL_OV_ODBC3, 0)
	if IsError(ret) {
		defer ReleaseHandle(Drv.H)
		return NewError("SQLSetEnvUIntPtrAttr", Drv.H)
	}
	return nil
}

// Close release the env handle
func (d *Driver) Close() error {
	// TODO(brainman): who will call (*Driver).Close (to dispose all opened handles)?
	h := d.H
	d.H = api.SQLHENV(api.SQL_NULL_HENV)
	return ReleaseHandle(h)
}

func init() {
	err := initDriver()
	if err != nil {
		panic(err)
	}
}
