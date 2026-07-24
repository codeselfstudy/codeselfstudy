package store

// UsernameMaxLen exposes the package-private width limit to the external
// store_test package, which asserts it still matches users.MaxLen. The store
// cannot import internal/users (that would be an import cycle), so the constant
// is duplicated — this is what keeps the copy honest.
const UsernameMaxLen = usernameMaxLen
