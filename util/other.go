package util

import (
  "strings"
  "strconv"
)

func FullNameToShort(fullName string) string {
  return strings.Split(fullName, "_")[2]
}

func ShortNameToFull(shortName string, t string, conNum int) string {
  return "eris_" + t + "_" + shortName + "_" + strconv.Itoa(conNum)
}