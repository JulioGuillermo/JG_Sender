package font

import (
	"fmt"

	"gioui.org/font/opentype"
	"gioui.org/text"
)

const (
	NotoSans           = "NotoSans"
	SauceCodeProNF     = "SauceCodeProNF"
	SauceCodeProNFCmp  = "SauceCodeProNFCmp"
	SauceCodeProNFMono = "SauceCodeProNFMono"
)

func GetNotoSans() text.FontFace {
	fnt := text.Font{
		Typeface: NotoSans,
	}
	face, err := opentype.Parse(otNotoSans)
	if err != nil {
		panic(fmt.Errorf("failed to parse font: %v", err))
	}
	return text.FontFace{Font: fnt, Face: face}
}

func GetNotoSansBold() text.FontFace {
	fnt := text.Font{
		Typeface: NotoSans,
		Weight:   text.Bold,
	}
	face, err := opentype.Parse(otNotoSansBold)
	if err != nil {
		panic(fmt.Errorf("failed to parse font: %v", err))
	}
	return text.FontFace{Font: fnt, Face: face}
}

func GetNotoSansItalic() text.FontFace {
	fnt := text.Font{
		Typeface: NotoSans,
		Style:    text.Italic,
	}
	face, err := opentype.Parse(otNotoSansItalic)
	if err != nil {
		panic(fmt.Errorf("failed to parse font: %v", err))
	}
	return text.FontFace{Font: fnt, Face: face}
}

func GetNotoSansBoldItalic() text.FontFace {
	fnt := text.Font{
		Typeface: NotoSans,
		Weight:   text.Bold,
		Style:    text.Italic,
	}
	face, err := opentype.Parse(otNotoSansBoldItalic)
	if err != nil {
		panic(fmt.Errorf("failed to parse font: %v", err))
	}
	return text.FontFace{Font: fnt, Face: face}
}

func GetSauceCodeProNF() text.FontFace {
	fnt := text.Font{
		Typeface: SauceCodeProNF,
	}
	face, err := opentype.Parse(otSauceCodeProNF)
	if err != nil {
		panic(fmt.Errorf("failed to parse font: %v", err))
	}
	return text.FontFace{Font: fnt, Face: face}
}

func GetSauceCodeProNFMono() text.FontFace {
	fnt := text.Font{
		Typeface: SauceCodeProNFMono,
	}
	face, err := opentype.Parse(otSauceCodeProNFMono)
	if err != nil {
		panic(fmt.Errorf("failed to parse font: %v", err))
	}
	return text.FontFace{Font: fnt, Face: face}
}

func GetSauceCodeProNFCmp() text.FontFace {
	fnt := text.Font{
		Typeface: SauceCodeProNFCmp,
	}
	face, err := opentype.Parse(otSauceCodeProNFCmp)
	if err != nil {
		panic(fmt.Errorf("failed to parse font: %v", err))
	}
	return text.FontFace{Font: fnt, Face: face}
}

func JGFonts() []text.FontFace {
	return []text.FontFace{
		GetNotoSans(),
		GetNotoSansBold(),
		GetNotoSansItalic(),
		GetNotoSansBoldItalic(),
		GetSauceCodeProNFMono(),
	}
}

func AppendFonts(fontCollection []text.FontFace) []text.FontFace {
	return append(fontCollection, JGFonts()...)
}
