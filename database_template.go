package databasetemplate

import (
    "database/sql"
	"fmt"
	"reflect"
)

type DatabaseTemplateImpl struct{
	Conn *sql.DB
}

type MapRow func(resultSet *sql.Rows)(object interface{},err error)

type DatabaseTemplate interface{
	Query(sql string,mapRow MapRow,params ...interface{})(object interface{},err error)
	Exec(sql string,params ...interface{})(err error)
	ExecForResult(sql string,params ...interface{})(sql.Result,error)
	QueryArray(sql string,mapRow MapRow,params ...interface{})([]interface{},error)
	QueryIntoArray(resultList interface{},sql string,mapRow MapRow,params ...interface{})(error)
	QueryObject(sql string,mapRow MapRow,params ...interface{})(object interface{},err error)
}

func (this *DatabaseTemplateImpl) Query(sql string,mapRow MapRow,params ...interface{})(object interface{},err error){
	result,error:=this.Conn.Query(sql,params...)
	defer result.Close()
	if error!=nil {
		err=error
		return
	}
	if result.Next(){
		object,err=mapRow(result)
	}else{
		return nil,nil
	}
	return
}

func (this *DatabaseTemplateImpl) QueryArray(sql string,mapRow MapRow,params ...interface{})([]interface{},error){
	result,err:=this.Conn.Query(sql,params...)
	defer result.Close()
	if err!=nil {
		return nil,err
	}
	var resArray []interface{}
	for result.Next(){
		obj,err:=mapRow(result)
		if err!=nil{
			return nil,err
		}
		resArray=append(resArray,obj)
	}
	return resArray,nil
}

func(this *DatabaseTemplateImpl) QueryIntoArray(resultList interface{},sql string,mapRow MapRow,params ...interface{})(error){
	pointerElements := true
	t, err := toType(resultList)
	if err != nil {
		var err2 error
		if t, err2 = toSliceType(resultList); t == nil {
			if err2 != nil {
				return err2
			}
			return  err
		}
		pointerElements = t.Kind() == reflect.Ptr
		if pointerElements {
			t = t.Elem()
		}
	}
	sliceValue := reflect.Indirect(reflect.ValueOf(resultList))
	result,err:=this.Conn.Query(sql,params...)
	defer result.Close()
	if err!=nil {
		return err
	}
	for result.Next(){
		v,err:=mapRow(result)
		if err!=nil{
			return err
		}
		if pointerElements {
			if reflect.TypeOf(v).Kind()==reflect.Ptr{
				value:=reflect.ValueOf(v)
				sliceValue.Set(reflect.Append(sliceValue,value))
			}else{
				return fmt.Errorf("can't get struct to pointer array")
			}
		}else{
			if reflect.TypeOf(v).Kind()!=reflect.Ptr{
				sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(v)))
			}else{
				return fmt.Errorf("can't get pointer to struct array")
			}
		}
	}
	if sliceValue.IsNil() {
		sliceValue.Set(reflect.MakeSlice(sliceValue.Type(), 0, 0))
	}
	return nil
}

func (this *DatabaseTemplateImpl) QueryObject(sql string,mapRow MapRow,params ...interface{})(object interface{},err error){
	result,error:=this.Conn.Query(sql,params...)
	defer result.Close()
	if error!=nil {
		err=error
		return
	}
	if result.Next(){
		object,err=mapRow(result)
	}
	return
}

func (this *DatabaseTemplateImpl) Exec(sql string,params ...interface{})(err error){
	_,error:=this.Conn.Exec(sql,params...)
	if error!=nil {
		err=error
	}
	return
}

func (this *DatabaseTemplateImpl) ExecForResult(sql string,params ...interface{})(sql.Result,error){
	result,error:=this.Conn.Exec(sql,params...)
	return result,error
}

// toSliceType returns the element type of the given object, if the object is a
// "*[]*Element" or "*[]Element". If not, returns nil.
// err is returned if the user was trying to pass a pointer-to-slice but failed.
func toSliceType(i interface{}) (reflect.Type, error) {
        t := reflect.TypeOf(i)
        if t.Kind() != reflect.Ptr {
                // If it's a slice, return a more helpful error message
                if t.Kind() == reflect.Slice {
                        return nil,fmt.Errorf("database_template: Cannot SELECT into a non-pointer slice: %v", t)
                }
                return nil, nil
        }
        if t = t.Elem(); t.Kind() != reflect.Slice {
                return nil, nil
        }
        return t.Elem(), nil
}

func toType(i interface{}) (reflect.Type, error) {
        t := reflect.TypeOf(i)

        // If a Pointer to a type, follow
        for t.Kind() == reflect.Ptr {
                t = t.Elem()
        }

        if t.Kind() != reflect.Struct {
                return nil, fmt.Errorf("database_template: Cannot SELECT into this type: %v", reflect.TypeOf(i))
        }
        return t, nil
}
