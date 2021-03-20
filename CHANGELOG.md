# v0.4.8 - 2021/3/21

### Reduce memory usage at compile time

* go-json have used about 2GB of memory at compile time, but now it can compile with about less than 550MB.

### Fix any encoder's bug

* Add many test cases for encoder
* Fix composite type ( slice/array/map )
* Fix pointer types
* Fix encoding of MarshalJSON or MarshalText or json.Number type

### Refactor encoder

* Change package layout for reducing memory usage at compile
* Remove anonymous and only operation
* Remove root property from encodeCompileContext and opcode

### Fix CI

* Add Go 1.16
* Remove Go 1.13
* Fix `make cover` task

### Number/Delim/Token/RawMessage use the types defined in encoding/json by type alias

# v0.4.7 - 2021/02/22

### Fix decoder

* Fix decoding of deep recursive structure
* Fix decoding of embedded unexported pointer field
* Fix invalid test case
* Fix decoding of invalid value
* Fix decoding of prefilled value
* Fix not being able to return UnmarshalTypeError when it should be returned
* Fix decoding of null value
* Fix decoding of type of null string
* Use pre allocated pointer if exists it at decoding

### Reduce memory usage at compile

* Integrate int/int8/int16/int32/int64 and uint/uint8/uint16/uint32/uint64 operation to reduce memory usage at compile

### Remove unnecessary optype
