package orgUtils

import "regexp"

var AccountRegex = regexp.MustCompile(`^\d{12}$`)
