package maileroo

import (
	"strings"
)

type EmailAddress struct {
	Address     string  `json:"address"`
	DisplayName *string `json:"display_name,omitempty"`
}

func NewEmail(address string, displayName string) EmailAddress {

	addr := strings.TrimSpace(address)

	var dn *string

	if strings.TrimSpace(displayName) != "" {
		name := strings.TrimSpace(displayName)
		dn = &name
	}

	return EmailAddress{Address: addr, DisplayName: dn}

}

func (e EmailAddress) ToJSON() map[string]string {

	if e.DisplayName == nil {

		return map[string]string{
			"address": e.Address,
		}

	}

	return map[string]string{
		"address":      e.Address,
		"display_name": *e.DisplayName,
	}

}

func emailAddressesToJSON(addrs []EmailAddress) []map[string]string {

	out := make([]map[string]string, 0, len(addrs))

	for _, a := range addrs {
		out = append(out, a.ToJSON())
	}

	return out

}
