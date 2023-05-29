package rt

import (
    `fmt`
    `reflect`
)

type Method struct {
    Id int
    Vt *GoType
}

func AsMethod(ty reflect.Type, id int) Method {
    return Method {
        Id: id,
        Vt: UnpackType(ty),
    }
}

func GetMethod(tp interface{}, name string) Method {
    if tp == nil {
        panic("value must be an interface pointer")
    } else if vt := reflect.TypeOf(tp); vt.Kind() != reflect.Ptr {
        panic("value must be an interface pointer")
    } else if et := vt.Elem(); et.Kind() != reflect.Interface {
        panic("value must be an interface pointer")
    } else if mm, ok := et.MethodByName(name); !ok {
        panic(fmt.Sprintf("interface %s does not have method %s", et.String(), name))
    } else {
        return AsMethod(et, mm.Index)
    }
}
