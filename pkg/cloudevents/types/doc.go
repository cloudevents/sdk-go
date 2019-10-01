/*
Package types implements the CloudEvents type system.

CloudEvents defines a small set of abstract types for all event context
attribute values, including custom extension attributes. Each CloudEvents type
has a corresponding native Go type and a canonical string encoding.

 +----------------+----------------+-----------------------------------+
 |CloudEvents Type|Native Type     |Convertible From Types             |
 +================+================+===================================+
 |Bool            |bool            |bool                               |
 +----------------+----------------+-----------------------------------+
 |Integer         |int32           |Any numeric type with value in     |
 |                |                |range of int32                     |
 +----------------+----------------+-----------------------------------+
 |String          |string          |string                             |
 +----------------+----------------+-----------------------------------+
 |Binary          |[]byte          |[]byte                             |
 +----------------+----------------+-----------------------------------+
 |URI-Reference   |*url.URL        |url.URL, types.URIRef, types.URI   |
 +----------------+----------------+-----------------------------------+
 |URI             |*url.URL        |url.URL, types.URIRef, types.URI   |
 |                |                |Must be an absolute URI.           |
 +----------------+----------------+-----------------------------------+
 |Timestamp       |time.Time       |time.Time, types.Timestamp         |
 +----------------+----------------+-----------------------------------+

The native types used to represent CloudEvents types are:

    bool, int32, string, []byte, *url.URL, time.Time

This package provides Parse... and Format... functions to convert to/from
canonical strings. Use standard url.URL.String() and url.Parse() for URL
values, the canonical string for a string is the string itself.

It provides To... functions to convert any extension attribute value, in native
or string form, to the expected type.

*/
package types
