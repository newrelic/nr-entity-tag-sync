package sync

import "strings"

func getNestedHelper(
  path []string,
  m map[string]interface{},
  index int,
) (string, bool) {
  if index >= len(path) {
    return "", false
  }

  v, ok := m[path[index]]
  if !ok {
    return "", false
  }

  switch u := v.(type) {
  case string:
    if index + 1 == len(path) {
      return u, true
    }
    return "", false

  case map[string]interface{}:
    return getNestedHelper(path, u, index + 1)

  default:
    if index + 1 == len(path) {
      return "", true
    }
    return "", false
  }
}

func getNestedKeyValue(
  path string,
  m map[string]interface{},
) (string, bool) {
  return getNestedHelper(strings.Split(path, "."), m, 0)
}

func getKeys(m map[string]string) []string {
  keys := make([]string, len(m))

  i := 0
  for k := range m {
    keys[i] = k
    i += 1
  }

  return keys
}

func stringSliceContains(slice []string, s string) bool {
  for _, v := range slice {
    if s == v {
      return true
    }
  }

  return false
}
