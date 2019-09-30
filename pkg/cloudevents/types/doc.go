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
 |URI-Reference   |url.URL         |url.URL, types.URIRef, types.URI   |
 +----------------+----------------+-----------------------------------+
 |URI             |url.URL         |url.URL, types.URIRef, types.URI   |
 |                |                |Must be an absolute URI.           |
 +----------------+----------------+-----------------------------------+
 |Timestamp       |time.Time       |time.Time, types.Timestamp         |
 +----------------+----------------+-----------------------------------+

NOTE: CloudEvents has overlapping types 'URI' and 'URI-Reference'. Both are
represented by the native url.URL

*/
package types
