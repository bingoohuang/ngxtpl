package ngxtpl_test

import (
	"testing"

	"github.com/bingoohuang/ngxtpl"
	"github.com/hashicorp/hcl"
	"github.com/stretchr/testify/assert"
)

func TestCfgParse(t *testing.T) {
	s, err := ngxtpl.ReadFileE("initassets/ngxtpl.hcl")
	assert.Nil(t, err)

	var cfg ngxtpl.Cfg

	assert.Nil(t, hcl.Unmarshal(s, &cfg))
}
