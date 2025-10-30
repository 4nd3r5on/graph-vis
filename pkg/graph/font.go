package graph

import "graph-positioner/pkg/common"

var (
	DefaultCharWidth  = 11 // pixels per character
	DefaultCharHeight = 18 // line height
)

type FontConfig struct {
	CharWidthPx  int `json:"charWidthPx"`
	CharHeightPx int `json:"charHeightPx"`
}

func FontConfigWithDefault(cfg FontConfig) FontConfig {
	return FontConfig{
		CharWidthPx:  common.UseNotZero(cfg.CharWidthPx, DefaultCharWidth),
		CharHeightPx: common.UseNotZero(cfg.CharHeightPx, DefaultCharHeight),
	}
}
