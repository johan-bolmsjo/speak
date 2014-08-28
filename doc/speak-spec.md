
Speak
=====

*Speak* is a lightweight interface definition language (IDL) similar to thrift
and protocol buffers. The encoding format is the one defined by
[msgpack](http://msgpack.org/).

Some defining features of *Speak*:

TODO: Rewrite this with some introductory piece.

- Source files are structured in packages.
- Choices select one of many other choice or message types.
- Messages contain tagged fields of basic, custom, choice or other
  message types.
- Message fields can be fixed or dynamic arrays of types.
- All message fields are optional.
- Message fields of basic types containing its zero value are not encoded.
- It's possible to define custom message field types that can be  extended
  with support functions in the native language.
- It's possible to decode encoded data without the IDL specification
  (although with some type information missing).
- With care it's possible to increase the size of integer message
  fields and stay compatible with old encoded data. The encoded
  integer format is always the smallest possible. Similarly array
  sizes can be increased and still be compatible.


Syntax
======

Most of the *Speak* syntax is defined in EBNF.
The following operators are used: Alteration `|`, grouping `()`, option `[]`, repetition `{}`.

Comments
--------

Comments start with the character sequence `//` and stop at the end of the line. 

Keywords
--------

The following words are keywords in *Speak*.

    choice    message
    end       package
    enum      type

Keywords are not allowed to be used as message field names.

Arrays
------

There are two types of arrays, fixed sized and dynamic sized. The syntax for
fixed sized arrays is "[*number*]", where *number* is a positive non-zero
integer value. The syntax for dynamic sized arrays is "[]".

    Array  = "[" [ PositiveNumber ] "]" .

Basic Types
-----------

    BasicType = "bool" | "byte" | "float32" | "float64" |
                "int8" | "int16" | "int32" | "int64" | "string" |
                "uint8" | "uint16" | "uint32" | "uint64" .

### bool

Boolean with value *true* or *false*, the zero value is *false*.

### float32, float64

32 and 64 bit IEEE floating point numbers, the zero value is *0*.

### int8, int16, int32, int64

Signed integers, the zero value is *0*.

### string

UTF8 string of dynamic length, the zero value is *""*.

### byte, uint8, uint16, uint32, uint64

Unsigned integers, the zero value is *0*.

Custom Types
------------

TODO:
Figure out how to make this work with choices.
Choices only work with "pointer" types.

Custom types can be used as message field types.
The intended purpose of custom types is to create blob like types, for example
IPv6 or SHA-1 types. They can only be created from basic types or arrays of
basic types.

Examples:

    type Ipv6Address [16]byte
    type Sha1        [20]byte

The grammar is as follows:

    TypeDef = "type" BigIdentifier [ Array ] BasicType NewLine .

Choices
-------

Choices select one of many other choice or message types.

    ChoiceDef        = "choice" BigIdentifier NewLine { ChoiceField } End .
    ChoiceField      = Tag FqTypeIdentifier NewLine .

Messages
--------

Messages contain tagged fields of basic, custom, choice or other
message types.

    MessageDef       = "message" BigIdentifier NewLine { MessageField } End .
    MessageField     = Tag LittleIdentifier [ Array ] MessageFieldType NewLine .
    MessageFieldType = BasicType | FqTypeIdentifier .

Enumerations
------------

Enumerations associate symbolic names with positive integer values. They are
encoded as msgpack integer types. The allowed value range is 0 to 2^32-1.

    EnumDef   = "enum" BigIdentifier NewLine { EnumField } End .
    EnumField = Tag BigIdentifier NewLine .

Packages
--------

Each *Speak* source file must belong to a package. A type can be referenced by
another package by prefixing the message field type with "packagename.". There
is no import statement as the referenced package name is available in the type
reference. Package dependencies must form a DAG.

    PackageDef = "package" Identifier NewLine .

Complete Grammar
----------------

The complete grammar to parse *Speak* (except comments).

    Grammar = { ChoiceDef | EnumDef | MessageDef | PackageDef | TypeDef } .

Misc Grammar
------------

    PositiveNumber   = Digit | "1" ... "9" { Digit } .
    Digit            = "0" ... "9" .
    Letter           = LowerCaseLetter | CapitalLetter .
    LowerCaseLetter  = "a" ... "z" .
    CapitalLetter    = "A" ... "Z" .
    Identifier       = Letter { Letter | Digit } .
    BigIdentifier    = CapitalLetter { Letter | Digit } .
    LittleIdentifier = LowerCaseLetter { Letter | Digit } .
    FqTypeIdentifier = [ Identifier "." ] BigIdentifier .
    Tag              = PositiveNumber ":" .
    End              = "end" NewLine .
    NewLine          = "\n" .


Encoding Format
===============

The encoding format is the one defined by [msgpack](http://msgpack.org/).
All data is encoded in network byte order.

Opcode Table
------------

Format: [Opcode] {Length | Value}[+] Type Comment

\+ Denotes that additional data follows the opcode and length field.
xxx Denotes bits in the opcode that hold the value or length.

    [0   xxxxxxx] {0}  Positive FixNum   // Range [0,127]
    [1 00 0 xxxx] {0}+ FixMap            // 0..15 entries
    [1 00 1 xxxx] {0}+ FixArray          // 0..15 entries
    [1 01 xxx xx] {0}+ FixRaw            // 0..31 bytes
    [1 10 000 10] {0}  false
    [1 10 000 11] {0}  true
    [1 10 010 10] {4}  float             // 32 bit IEEE float
    [1 10 010 11] {8}  double            // 64 bit IEEE float
    [1 10 011 00] {1}  uint 8
    [1 10 011 01] {2}  uint 16
    [1 10 011 10] {4}  uint 32
    [1 10 011 11] {8}  uint 64
    [1 10 100 00] {1}  int 8
    [1 10 100 01] {2}  int 16
    [1 10 100 10] {4}  int 32
    [1 10 100 11] {8}  int 64
    [1 10 110 10] {2}+ raw 16
    [1 10 110 11] {4}+ raw 32
    [1 10 111 00] {2}+ array 16
    [1 10 111 01] {4}+ array 32
    [1 10 111 10] {2}+ map 16
    [1 10 111 11] {4}+ map 32
    [1 11 xxx xx] {0}  Negative FixNum   // Range [-32,-1]

Arrays
------

Arrays are used to encode *Speak* arrays except byte arrays.
The length specifies the number of elements that follows.

Integers
--------

The *Speak* integer types are encoded to the smallest possible msgpack integer
type.

Maps
----

Maps are used to encode *Speak* messages and choices. Tags are keys and fields
values. The length specifies the number of key + value pairs that follows.

Raw (Bytes)
-----------

Raw buffers are used to encode *Speak* byte arrays and strings. The length
specifies the number of bytes that follows.


Example
=======

    package paint

    enum Color
        1: Red
        2: Green
        3: Blue
    end

    type XyCoordinate [2]float32

    message PaintRequest
        1: id           msg.Id        // Use type 'Id' in package 'msg'.
        2: color        Color
        3: brushSize    float32       // Brush size in millimetres.
        4: xyCoordinate XyCoordinate
    end


Design Considerations
=====================

The rationale for some design choices are listed here. The primary languages
that code generation is planed for is C, Go and C++, this affects some
decisions.

- No underscore letters are allowed in the field or type names to be able to
  string together package, type and field names with underscores in the C
  generated code. Camel case will have to do.

- The rule that types must begin with a capital letter and fields with a
  lowercase one is mostly for aesthetic reasons. However in Go the field names
  will have to be converted to begin with a capital letter to be exported from
  its package. At least this rule will disallow two fields with the same name
  except the capitalisation of the first letter.

- Package and type dependencies must form a DAG to make code generation simpler.


Implementation Hints
====================

Name collisions with target language keywords
---------------------------------------------

If a message field name happens to be a keyword in the target language
the code generator can choose a slightly altered name. For example
when generating code for C, a message field name called *int* could be
renamed to *Int* in the generated code.
