package sum

import (
	"testing"
)

func TestParseBmcInfo(t *testing.T) {
	docs := []string{
		// sum -i <HOST> -u ADMIN -p <pass> -c GetBmcInfo
		`
Supermicro Update Manager (for UEFI BIOS) 2.10.0 (2022/12/09) (x86_64)
Copyright(C) 2013-2022 Super Micro Computer, Inc. All rights reserved.
............
Managed system...........lsstcam-dc01-bmc.ls.lsst.org
    BMC UFFN.............BMC_H12AST2500-ROT-2201MS_20230106_01.01.08_STDsp.bin
    BMC type.............H12_RoT_ATEN_AST2500/H12_RoT_ATEN_AST2600_1_2
    BMC version..........01.01.08
    BMC ext. version.....01 00 00 (P)
    BMC build date.......2023/01/06
`,
		// sum -c GetBmcInfo --file BMC_H12AST2500-ROT_2201MS_20230106_01.01.08sp.bin --file_only
		`
Supermicro Update Manager (for UEFI BIOS) 2.10.0 (2022/12/09) (x86_64)
Copyright(C) 2013-2022 Super Micro Computer, Inc. All rights reserved.

Local BMC image file...../home/jhoblitt/Dropbox/lsst-it/sm/AS-1114S-WN10RT/bmc/BMC_H12AST2500-ROT_2201MS_20230106_01.01.08sp.bin
    BMC UFFN.............BMC_H12AST2500-ROT-2201MS_20230106_01.01.08_STDsp.bin
    BMC type.............H12_RoT_ATEN_AST2500/H12_RoT_ATEN_AST2600_1_2
    BMC version..........01.01.08
    BMC build date.......2023/01/06
    FW image.............Signed
        Signed Key.......RoT
`}

	for _, doc := range docs {
		bmc, err := ParseBmcInfo(doc)
		if err != nil {
			t.Fatalf("unexpected error parsing BMC info: %s", err)
		}
		if bmc.UFFN != "BMC_H12AST2500-ROT-2201MS_20230106_01.01.08_STDsp.bin" {
			t.Fatalf("unexpected BMC UFFN %q", bmc.UFFN)
		}
		if bmc.Type != "H12_RoT_ATEN_AST2500/H12_RoT_ATEN_AST2600_1_2" {
			t.Fatalf("unexpected BMC type %q", bmc.Type)
		}
		if bmc.Version != "01.01.08" {
			t.Fatalf("unexpected BMC version %q", bmc.Version)
		}
		if bmc.Date != "2023/01/06" {
			t.Fatalf("unexpected BMC build date %q", bmc.Date)
		}
	}
}

func TestParseBiosInfo(t *testing.T) {
	docs := []string{
		// sum -i <HOST> -u ADMIN -p <pass> -c GetBiosInfo
		`
Supermicro Update Manager (for UEFI BIOS) 2.11.0 (2023/05/11) (x86_64)
Copyright(C) 2013-2023 Super Micro Computer, Inc. All rights reserved.
.........
Managed system..........................core01-bmc.dev.lsst.org
    Board ID............................1B74
    BIOS build date.....................2022/09/23
`,
		// sum -c GetBiosInfo --file /home/jhoblitt/Dropbox/lsst-it/sm/AS-1114S-WN10RT/bios/2.5/BIOS_H12SSWR-1B74_20220923_2.5_STDsp.bin --file_only
		`
Supermicro Update Manager (for UEFI BIOS) 2.10.0 (2022/12/09) (x86_64)
Copyright(C) 2013-2022 Super Micro Computer, Inc. All rights reserved.

Local BIOS image file (Rome)............/home/jhoblitt/Dropbox/lsst-it/sm/AS-1114S-WN10RT/bios/2.5/BIOS_H12SSWR-1B74_20220923_2.5_STDsp.bin
    Board ID............................1B74
    BIOS build date.....................2022/09/23
    FW image............................Signed
        Signed Key......................RoT

Local BIOS image file (Milan).........../home/jhoblitt/Dropbox/lsst-it/sm/AS-1114S-WN10RT/bios/2.5/BIOS_H12SSWR-1B74_20220923_2.5_STDsp.bin
    Board ID............................1B74
    BIOS build date.....................2022/09/23
    FW image............................Signed
        Signed Key......................RoT
`}

	for _, doc := range docs {
		bios, err := ParseBiosInfo(doc)
		if err != nil {
			t.Fatalf("unexpected error parsing BIOS info: %s", err)
		}
		if bios.BoardID != "1B74" {
			t.Fatalf("unexpected Board ID %q", bios.BoardID)
		}
		if bios.BuildDate != "2022/09/23" {
			t.Fatalf("unexpected BIOS build date %q", bios.BuildDate)
		}
	}
}

func TestNewSum(t *testing.T) {
	sum := NewSum("/does/not/exist")
	if sum == nil {
		t.Fatalf("unexpected failure in New()")
	}
}
