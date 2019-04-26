# exprel [![GoDoc](https://godoc.org/layeh.com/exprel?status.svg)](https://godoc.org/layeh.com/exprel)

Package exprel provides a Spreadsheet-like expression evaluator.

Example expressions which exprel can evaluate:

    Expression                        Return value
    ----------------------------------------------
    Hey there                         "Hey there"
    1234                              "1234"
    =5+5*2                            15
    ="A" & " " & "B"                  "A B"
    =IF(AND(NOT(FALSE());1=1);1+2;2)  3

## Documentation

Documentation is provided through the package's godoc. The documentation can be viewed online at [godoc.org](https://godoc.org/layeh.com/exprel).

## License

[MPL 2.0](https://www.mozilla.org/en-US/MPL/2.0/)

## Author

Tim Cooper (<tim.cooper@layeh.com>)
