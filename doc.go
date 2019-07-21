// Package odb provides an registry for an O.S. to be shared between
// multiple applications.
//
// Each entry can be set using a path information and versions are kept,
// so updates works like a stack instead of replacing the data.
//
// At some point old versions can be discarded.
//
// To read/write to a node the node must be public or the caller must
// have an authorized token.
package odb
