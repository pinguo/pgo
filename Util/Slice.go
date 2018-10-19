package Util

// SliceSearch search in a slice has the length of n,
// return the first position where f(i) is true,
// -1 is returned if nothing found.
func SliceSearch(n int, f func(int) bool) int {
    for i := 0; i < n; i++ {
        if f(i) {
            return i
        }
    }

    return -1
}

// SliceSearchInt search x in an int slice, return the first position of x,
// -1 is returned if nothing found.
func SliceSearchInt(a []int, x int) int {
    return SliceSearch(len(a), func(i int) bool { return a[i] == x })
}

// SliceSearchFloat search x in a float64 slice, return the first position of x,
// -1 is returned if nothing found.
func SliceSearchFloat(a []float64, x float64) int {
    return SliceSearch(len(a), func(i int) bool { return a[i] == x })
}

// SliceSearchString search x in a string slice, return the first position of x,
// -1 is returned if nothing found.
func SliceSearchString(a []string, x string) int {
    return SliceSearch(len(a), func(i int) bool { return a[i] == x })
}
