/*
Package types implements the CloudEvents type system.

The CloudEvents specifics defines a set of abstract types for all event context
attribute values, including custom extension attributes. Each CloudEvents type
has a corresponding Go type and a standard string encoding. Event.Extensions()
values can be in either form.

The To...() functions in this package convert to the Go representation from any
convertible Go type or the CloudEvents string representation. The StringOf()
function converts any value to the corresponding CloudEvents string form. The
Normalize() function converts convertible values to the "normal" type, for
example it converts any legal numeric type to int32.


+----------------+---------+-----------------------------------+
|CloudEvents Type|Go type  |Compatible types                   |
+----------------+---------+-----------------------------------+
|Bool            |bool     |bool                               |
+----------------+---------+-----------------------------------+
|Integer         |int32    |Any numeric type with value in     |
|                |         |range of int32                     |
+----------------+---------+-----------------------------------+
|String          |string   |string                             |
+----------------+---------+-----------------------------------+
|Binary          |[]byte   |[]byte                             |
+----------------+---------+-----------------------------------+
|URI             |url.URL  |url.URL, types.URIRef.             |
|                |         |Must be an absolute URI.           |
+----------------+---------+-----------------------------------+
|URI-Reference   |url.URL  |url.URL, types.URIRef              |
+----------------+---------+-----------------------------------+
|Timestamp       |time.Time|time.Time, types.Timestamp         |
+----------------+---------+-----------------------------------+

This package also provides some wrapper types (e.g. type.URIRef) to ensure
correct marshalling when used as struct members.

*/
package types
