# fproto-wrap-validator - Golang

[![GoDoc](https://godoc.org/github.com/RangelReale/fproto-wrap-validator/gowrap?status.svg)](https://godoc.org/github.com/RangelReale/fproto-wrap-validator/gowrap)

Package for generating validators to [fproto-wrap](https://github.com/RangelReale/fproto-wrap/tree/master/gowrap) generated structs and interfaces.

For each "message" and "oneof" which contains validations, this method is added:

    func (m *Message) Validate() error

The returned error is *always* of the type "github.com/RangelReale/fproto-wrap-validator/gowrap/runtime".*Error. From
this error type you can extract all types of information which can be helpful to generate your own error messages and codes.

All validation definitions are created by plugins, this package itself doesn't define any validators.

It is also possible to generate custom validation for determined types using type validators, like time, uuid, etc.

### example

This example uses the [Std](https://github.com/RangelReale/fproto-wrap-validator-std) and
[Govalidator](https://github.com/RangelReale/fproto-wrap-validator-govalidator) validator plugins.

```protobuf
syntax = "proto3";
package gw_sample;
option go_package = "gwsample/core";

import "github.com/RangelReale/fproto-wrap-validator-std/validator.proto";
import "github.com/RangelReale/fproto-wrap-validator-govalidator/govalidator.proto";

message User {
    string name = 1 [(validator.field) = {required: true}];
    string email = 2 [(validator.field) = {required: true}, (govalidator.field) = {email: true}]
}
```

And in the wrapper generation program:

```go
    // create the fdep parser
    parsedep := fdep.NewDep()

    // ... add files here ...

    // create the fproto-wrap wrapper
    w := fproto_gowrap.NewWrapper(parsedep)

    // ... configure wrapper here ...

    // create the core validator
    val := fproto_gowrap_validator.NewCustomizer_Validator()

    // Add both validators to the core validator
    val.Validators = append(val.Validators,
        &fproto_gowrap_validator_std.ValidatorPlugin_Std{},
        &fproto_gowrap_validator_govalidator.ValidatorPlugin_Govalidator{},
    )

    // Add some type validators if needed
    val.TypeValidators = append(val.TypeValidators,
        &fproto_gowrap_validator_std_uuid.TypeValidatorPlugin_UUID{},
        &fproto_gowrap_validator_std_time.TypeValidatorPlugin_Time{},
        &fproto_gowrap_validator_std_duration.TypeValidatorPlugin_Duration{},
    )

	// Add the core validator to the fproto-wrap customizers
    w.Customizers = append(w.Customizers,
        val,
    )
```

### validations

* Std - [https://github.com/RangelReale/fproto-wrap-validator-std](https://github.com/RangelReale/fproto-wrap-validator-std) (standard validators like "required", "int_gt", "float_gte", "length_eq", etc)
* Govalidator - [https://github.com/RangelReale/fproto-wrap-validator-govalidator](https://github.com/RangelReale/fproto-wrap-validator-govalidator) (validations using [govalidator](https://github.com/asaskevich/govalidator))

### author

Rangel Reale (rangelspam@gmail.com)
