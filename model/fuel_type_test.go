package model

import (
	"math"
	"testing"
)

// The values below are the ones the IMO publishes, and the expected results are
// worked out by hand from them, so a change in the formulas has to be a
// deliberate change rather than a silent drift.
//
//	MEPC.364(79) Annex 9      carbon factors and lower heating values
//	MEPC.391(81) Appendix 2   well-to-tank intensities, CH4/N2O, methane slip
//	Fourth IMO GHG Study 2020 Table 10  densities
func ptr(v float64) *float64 { return &v }

func TestTtWCO2KgPerLitre(t *testing.T) {
	// Marine gas oil: 3.206 kg CO2 per kg fuel x 0.895 kg per litre.
	mgo := MFuelType{CfCO2KgPerKg: ptr(3.206), DensityKgPerLitre: ptr(0.895)}
	got, ok := mgo.TtWCO2KgPerLitre()
	if !ok {
		t.Fatal("expected the factor to be derivable")
	}
	if math.Abs(got-2.86937) > 1e-6 {
		t.Errorf("ttw factor = %v, want 2.86937", got)
	}

	// Heavy fuel oil: 3.114 x 1.001.
	hfo := MFuelType{CfCO2KgPerKg: ptr(3.114), DensityKgPerLitre: ptr(1.001)}
	got, ok = hfo.TtWCO2KgPerLitre()
	if !ok {
		t.Fatal("expected the factor to be derivable")
	}
	if math.Abs(got-3.117114) > 1e-6 {
		t.Errorf("ttw factor = %v, want 3.117114", got)
	}

	// Liquefied natural gas: 2.750 x 0.450.
	lng := MFuelType{CfCO2KgPerKg: ptr(2.750), DensityKgPerLitre: ptr(0.450)}
	got, ok = lng.TtWCO2KgPerLitre()
	if !ok {
		t.Fatal("expected the factor to be derivable")
	}
	if math.Abs(got-1.2375) > 1e-9 {
		t.Errorf("ttw factor = %v, want 1.2375", got)
	}
}

func TestFactorsNeedBothTheCarbonFactorAndTheDensity(t *testing.T) {
	// A carbon factor without a density cannot be turned into a figure per
	// litre, and a density without a carbon factor says nothing.
	noDensity := MFuelType{CfCO2KgPerKg: ptr(3.206)}
	if _, ok := noDensity.TtWCO2KgPerLitre(); ok {
		t.Error("expected no factor without a density")
	}
	noCarbon := MFuelType{DensityKgPerLitre: ptr(0.895)}
	if _, ok := noCarbon.TtWCO2KgPerLitre(); ok {
		t.Error("expected no factor without a carbon factor")
	}
}

func TestTtWCO2eIncludesTheGreenhouseGasesAndTheMethaneSlip(t *testing.T) {
	// Marine gas oil: 3.206 + 28*0.00005 + 265*0.00018 = 3.25510 per kg.
	mgo := MFuelType{
		CfCO2KgPerKg:      ptr(3.206),
		CfCH4KgPerKg:      ptr(0.00005),
		CfN2OKgPerKg:      ptr(0.00018),
		DensityKgPerLitre: ptr(0.895),
	}
	perKg, ok := mgo.TtWCO2eKgPerKg()
	if !ok {
		t.Fatal("expected the factor to be derivable")
	}
	if math.Abs(perKg-3.2551) > 1e-9 {
		t.Errorf("ttw co2e = %v, want 3.2551", perKg)
	}

	// Liquefied natural gas with a 3.5 % methane slip: the slipped methane is
	// released unburned, so it adds 28 * 0.035 = 0.98 per kg.
	lng := MFuelType{
		CfCO2KgPerKg:      ptr(2.750),
		CfN2OKgPerKg:      ptr(0.00011),
		MethaneSlipPc:     ptr(3.5),
		DensityKgPerLitre: ptr(0.450),
	}
	perKg, ok = lng.TtWCO2eKgPerKg()
	if !ok {
		t.Fatal("expected the factor to be derivable")
	}
	want := 2.750 + 265*0.00011 + 28*0.035
	if math.Abs(perKg-want) > 1e-9 {
		t.Errorf("ttw co2e = %v, want %v", perKg, want)
	}
}

func TestWtWNeedsAWellToTankIntensity(t *testing.T) {
	// Marine gas oil: WtT 17.7 gCO2eq/MJ x LCV 42.7 MJ/kg = 0.75579 per kg,
	// plus the tank-to-wake part.
	mgo := MFuelType{
		CfCO2KgPerKg:            ptr(3.206),
		CfCH4KgPerKg:            ptr(0.00005),
		CfN2OKgPerKg:            ptr(0.00018),
		WttIntensityGCO2eqPerMJ: ptr(17.7),
		LhvMJPerKg:              ptr(42.7),
		DensityKgPerLitre:       ptr(0.895),
	}
	perKg, ok := mgo.WtWCO2eKgPerKg()
	if !ok {
		t.Fatal("expected the factor to be derivable")
	}
	want := 17.7*42.7/1000 + 3.2551
	if math.Abs(perKg-want) > 1e-9 {
		t.Errorf("wtw co2e = %v, want %v", perKg, want)
	}

	// The IMO has no well-to-tank default for every fuel. Without one the
	// well-to-wake figure is not derivable, and it has to be reported as
	// unavailable rather than guessed.
	lng := MFuelType{
		CfCO2KgPerKg:      ptr(2.750),
		DensityKgPerLitre: ptr(0.450),
	}
	if _, ok := lng.WtWCO2eKgPerKg(); ok {
		t.Error("expected no well-to-wake factor without a well-to-tank intensity")
	}
	if _, ok := lng.WtWCO2eKgPerLitre(); ok {
		t.Error("expected no well-to-wake factor without a well-to-tank intensity")
	}
}
