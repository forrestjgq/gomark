package gm

type Operator func(left, right Value) Value
type OperatorInt func(left Value, right int) Value
