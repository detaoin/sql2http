// +build oci8

// This file builds only with the specific tag "oci8" because it needs
// the Oracle SQL SDK at compile time.

package main

import _ "github.com/mattn/go-oci8"
