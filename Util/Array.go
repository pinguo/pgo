package Util

import "sort"

func MergeString(a []string, b []string) ([]string) {
    for _, v := range (b) {
        a = append(a, v)
    }
    return a
}

//Int returns unique int values in a slice
func UniqueInt(slice []int) []int {
    uMap := make(map[int]struct{})
    result := []int{}
    for _, val := range slice {
        uMap[val] = struct{}{}
    }
    for key := range uMap {
        result = append(result, key)
    }
    sort.Ints(result)
    return result
}

//Strings returns unique string values in a slice
func UniqueStrings(slice []string) []string {
    uMap := make(map[string]struct{})
    result := []string{}
    for _, val := range slice {
        uMap[val] = struct{}{}
    }
    for key := range uMap {
        result = append(result, key)
    }
    sort.Strings(result)
    return result
}
