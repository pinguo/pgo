package Util

// search in a slice has the length of n, return the
// first position where f(i) is true, -1 is returned
// if nothing found.
func SliceSearch(n int, f func(int) bool) int {
    for i := 0; i < n; i++ {
        if f(i) {
            return i
        }
    }

    return -1
}

// search x in an int slice, return the first position of x,
// -1 is returned if nothing found.
func SliceSearchInt(a []int, x int) int {
    return SliceSearch(len(a), func(i int) bool { return a[i] == x })
}

// search x in a float64 slice, return the first position of x,
// -1 is returned if nothing found.
func SliceSearchFloat(a []float64, x float64) int {
    return SliceSearch(len(a), func(i int) bool { return a[i] == x })
}

// search x in a string slice, return the first position of x,
// -1 is returned if nothing found.
func SliceSearchString(a []string, x string) int {
    return SliceSearch(len(a), func(i int) bool { return a[i] == x })
}

// 是否在数组里面
// need  需要查找的字符串
// 被查找的数组
func InArrayStr(need string, haystack []string) bool {
    for _, v := range haystack {
        if v == need {
            return true
        }
    }
    return false
}

func ArrayEqualStr(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }

    if (a == nil) != (b == nil) {
        return false
    }

    for i, v := range a {
        if v != b[i] {
            return false
        }
    }

    return true
}

func ArrayIntersectStr(a, b []string) []string {
    var ret []string
    for _, v := range a {
        if InArrayStr(v, b) {
            ret = append(ret, v)
        }
    }
    return ret
}

func ArrayDiffStr(a, b []string) []string {
    var ret []string
    for _, v := range a {
        if InArrayStr(v, b) == false {
            ret = append(ret, v)
        }
    }
    return ret
}
