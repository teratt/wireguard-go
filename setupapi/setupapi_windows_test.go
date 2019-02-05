/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package setupapi

import (
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/windows"
)

var deviceClassNetGUID = windows.GUID{0x4d36e972, 0xe325, 0x11ce, [8]byte{0xbf, 0xc1, 0x08, 0x00, 0x2b, 0xe1, 0x03, 0x18}}
var computerName string

func init() {
	computerName, _ = windows.ComputerName()
}

func TestSetupDiCreateDeviceInfoListEx(t *testing.T) {
	devInfoList, err := SetupDiCreateDeviceInfoListEx(&deviceClassNetGUID, 0, "")
	if err == nil {
		devInfoList.Close()
	} else {
		t.Errorf("Error calling SetupDiCreateDeviceInfoListEx: %s", err.Error())
	}

	devInfoList, err = SetupDiCreateDeviceInfoListEx(&deviceClassNetGUID, 0, computerName)
	if err == nil {
		devInfoList.Close()
	} else {
		t.Errorf("Error calling SetupDiCreateDeviceInfoListEx: %s", err.Error())
	}

	devInfoList, err = SetupDiCreateDeviceInfoListEx(nil, 0, "")
	if err == nil {
		devInfoList.Close()
	} else {
		t.Errorf("Error calling SetupDiCreateDeviceInfoListEx(nil): %s", err.Error())
	}
}

func TestSetupDiGetDeviceInfoListDetail(t *testing.T) {
	devInfoList, err := SetupDiGetClassDevsEx(&deviceClassNetGUID, "", 0, DIGCF_PRESENT, DevInfo(0), "")
	if err != nil {
		t.Errorf("Error calling SetupDiGetClassDevsEx: %s", err.Error())
	}
	defer devInfoList.Close()

	data, err := devInfoList.GetDeviceInfoListDetail()
	if err != nil {
		t.Errorf("Error calling SetupDiGetDeviceInfoListDetail: %s", err.Error())
	} else {
		if data.ClassGUID != deviceClassNetGUID {
			t.Error("SetupDiGetDeviceInfoListDetail returned different class GUID")
		}

		if data.RemoteMachineHandle != windows.Handle(0) {
			t.Error("SetupDiGetDeviceInfoListDetail returned non-NULL remote machine handle")
		}

		if data.RemoteMachineName != "" {
			t.Error("SetupDiGetDeviceInfoListDetail returned non-NULL remote machine name")
		}
	}

	devInfoList, err = SetupDiGetClassDevsEx(&deviceClassNetGUID, "", 0, DIGCF_PRESENT, DevInfo(0), computerName)
	if err != nil {
		t.Errorf("Error calling SetupDiGetClassDevsEx: %s", err.Error())
	}
	defer devInfoList.Close()

	data, err = devInfoList.GetDeviceInfoListDetail()
	if err != nil {
		t.Errorf("Error calling SetupDiGetDeviceInfoListDetail: %s", err.Error())
	} else {
		if data.ClassGUID != deviceClassNetGUID {
			t.Error("SetupDiGetDeviceInfoListDetail returned different class GUID")
		}

		if data.RemoteMachineHandle == windows.Handle(0) {
			t.Error("SetupDiGetDeviceInfoListDetail returned NULL remote machine handle")
		}

		if data.RemoteMachineName != computerName {
			t.Error("SetupDiGetDeviceInfoListDetail returned different remote machine name")
		}
	}
}

func TestSetupDiCreateDeviceInfo(t *testing.T) {
	devInfoList, err := SetupDiCreateDeviceInfoListEx(&deviceClassNetGUID, 0, computerName)
	if err != nil {
		t.Errorf("Error calling SetupDiCreateDeviceInfoListEx: %s", err.Error())
	}
	defer devInfoList.Close()

	deviceClassNetName, err := SetupDiClassNameFromGuidEx(&deviceClassNetGUID, computerName)
	if err != nil {
		t.Errorf("Error calling SetupDiClassNameFromGuidEx: %s", err.Error())
	}

	devInfoData, err := devInfoList.CreateDeviceInfo(deviceClassNetName, &deviceClassNetGUID, "This is a test device", 0, DICD_GENERATE_ID)
	if err != nil {
		// Access denied is expected, as the SetupDiCreateDeviceInfo() require elevation to succeed.
		if errWin, ok := err.(syscall.Errno); !ok || errWin != windows.ERROR_ACCESS_DENIED {
			t.Errorf("Error calling SetupDiCreateDeviceInfo: %s", err.Error())
		}
	} else if devInfoData.ClassGUID != deviceClassNetGUID {
		t.Error("SetupDiCreateDeviceInfo returned different class GUID")
	}
}

func TestSetupDiEnumDeviceInfo(t *testing.T) {
	devInfoList, err := SetupDiGetClassDevsEx(&deviceClassNetGUID, "", 0, DIGCF_PRESENT, DevInfo(0), "")
	if err != nil {
		t.Errorf("Error calling SetupDiGetClassDevsEx: %s", err.Error())
	}
	defer devInfoList.Close()

	for i := 0; true; i++ {
		data, err := devInfoList.EnumDeviceInfo(i)
		if err != nil {
			if errWin, ok := err.(syscall.Errno); ok && errWin == 259 /*ERROR_NO_MORE_ITEMS*/ {
				break
			}
			continue
		}

		if data.ClassGUID != deviceClassNetGUID {
			t.Error("SetupDiEnumDeviceInfo returned different class GUID")
		}
	}
}

func TestSetupDiGetClassDevsEx(t *testing.T) {
	devInfoList, err := SetupDiGetClassDevsEx(&deviceClassNetGUID, "PCI", 0, DIGCF_PRESENT, DevInfo(0), computerName)
	if err == nil {
		devInfoList.Close()
	} else {
		t.Errorf("Error calling SetupDiGetClassDevsEx: %s", err.Error())
	}

	devInfoList, err = SetupDiGetClassDevsEx(nil, "", 0, DIGCF_PRESENT, DevInfo(0), "")
	if err == nil {
		devInfoList.Close()
		t.Errorf("SetupDiGetClassDevsEx(nil, ...) should fail")
	} else {
		if errWin, ok := err.(syscall.Errno); !ok || errWin != 87 /*ERROR_INVALID_PARAMETER*/ {
			t.Errorf("SetupDiGetClassDevsEx(nil, ...) should fail with ERROR_INVALID_PARAMETER")
		}
	}
}

func TestSetupDiOpenDevRegKey(t *testing.T) {
	devInfoList, err := SetupDiGetClassDevsEx(&deviceClassNetGUID, "", 0, DIGCF_PRESENT, DevInfo(0), "")
	if err != nil {
		t.Errorf("Error calling SetupDiGetClassDevsEx: %s", err.Error())
	}
	defer devInfoList.Close()

	for i := 0; true; i++ {
		data, err := devInfoList.EnumDeviceInfo(i)
		if err != nil {
			if errWin, ok := err.(syscall.Errno); ok && errWin == 259 /*ERROR_NO_MORE_ITEMS*/ {
				break
			}
			continue
		}

		key, err := devInfoList.OpenDevRegKey(data, DICS_FLAG_GLOBAL, 0, DIREG_DRV, windows.KEY_READ)
		if err != nil {
			t.Errorf("Error calling SetupDiOpenDevRegKey: %s", err.Error())
		}
		defer key.Close()
	}
}

func TestSetupDiGetDeviceInstallParams(t *testing.T) {
	devInfoList, err := SetupDiGetClassDevsEx(&deviceClassNetGUID, "", 0, DIGCF_PRESENT, DevInfo(0), "")
	if err != nil {
		t.Errorf("Error calling SetupDiGetClassDevsEx: %s", err.Error())
	}
	defer devInfoList.Close()

	for i := 0; true; i++ {
		data, err := devInfoList.EnumDeviceInfo(i)
		if err != nil {
			if errWin, ok := err.(syscall.Errno); ok && errWin == 259 /*ERROR_NO_MORE_ITEMS*/ {
				break
			}
			continue
		}

		_, err = devInfoList.GetDeviceInstallParams(data)
		if err != nil {
			t.Errorf("Error calling SetupDiGetDeviceInstallParams: %s", err.Error())
		}
	}
}

func TestSetupDiClassNameFromGuidEx(t *testing.T) {
	deviceClassNetName, err := SetupDiClassNameFromGuidEx(&deviceClassNetGUID, "")
	if err != nil {
		t.Errorf("Error calling SetupDiClassNameFromGuidEx: %s", err.Error())
	} else if strings.ToLower(deviceClassNetName) != "net" {
		t.Errorf("SetupDiClassNameFromGuidEx(%x) should return \"Net\"", deviceClassNetGUID)
	}

	deviceClassNetName, err = SetupDiClassNameFromGuidEx(&deviceClassNetGUID, computerName)
	if err != nil {
		t.Errorf("Error calling SetupDiClassNameFromGuidEx: %s", err.Error())
	} else if strings.ToLower(deviceClassNetName) != "net" {
		t.Errorf("SetupDiClassNameFromGuidEx(%x) should return \"Net\"", deviceClassNetGUID)
	}

	_, err = SetupDiClassNameFromGuidEx(nil, "")
	if err == nil {
		t.Errorf("SetupDiClassNameFromGuidEx(nil) should fail")
	} else {
		if errWin, ok := err.(syscall.Errno); !ok || errWin != 1784 /*ERROR_INVALID_USER_BUFFER*/ {
			t.Errorf("SetupDiClassNameFromGuidEx(nil) should fail with ERROR_INVALID_USER_BUFFER")
		}
	}
}

func TestSetupDiClassGuidsFromNameEx(t *testing.T) {
	ClassGUIDs, err := SetupDiClassGuidsFromNameEx("Net", "")
	if err != nil {
		t.Errorf("Error calling SetupDiClassGuidsFromNameEx: %s", err.Error())
	} else {
		found := false
		for i := range ClassGUIDs {
			if ClassGUIDs[i] == deviceClassNetGUID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SetupDiClassGuidsFromNameEx(\"Net\") should return %x", deviceClassNetGUID)
		}
	}

	ClassGUIDs, err = SetupDiClassGuidsFromNameEx("foobar-34274a51-a6e6-45f0-80d6-c62be96dd5fe", computerName)
	if err != nil {
		t.Errorf("Error calling SetupDiClassGuidsFromNameEx: %s", err.Error())
	} else if len(ClassGUIDs) != 0 {
		t.Errorf("SetupDiClassGuidsFromNameEx(\"foobar-34274a51-a6e6-45f0-80d6-c62be96dd5fe\") should return an empty GUID set")
	}
}

func TestSetupDiGetSelectedDevice(t *testing.T) {
	devInfoList, err := SetupDiGetClassDevsEx(&deviceClassNetGUID, "", 0, DIGCF_PRESENT, DevInfo(0), "")
	if err != nil {
		t.Errorf("Error calling SetupDiGetClassDevsEx: %s", err.Error())
	}
	defer devInfoList.Close()

	for i := 0; true; i++ {
		data, err := devInfoList.EnumDeviceInfo(i)
		if err != nil {
			if errWin, ok := err.(syscall.Errno); ok && errWin == 259 /*ERROR_NO_MORE_ITEMS*/ {
				break
			}
			continue
		}

		err = devInfoList.SetSelectedDevice(data)
		if err != nil {
			t.Errorf("Error calling SetupDiSetSelectedDevice: %s", err.Error())
		}

		data2, err := devInfoList.GetSelectedDevice()
		if err != nil {
			t.Errorf("Error calling SetupDiGetSelectedDevice: %s", err.Error())
		} else if *data != *data2 {
			t.Error("SetupDiGetSelectedDevice returned different data than was set by SetupDiSetSelectedDevice")
		}
	}

	err = devInfoList.SetSelectedDevice(nil)
	if err == nil {
		t.Errorf("SetupDiSetSelectedDevice(nil) should fail")
	} else {
		if errWin, ok := err.(syscall.Errno); !ok || errWin != 87 /*ERROR_INVALID_PARAMETER*/ {
			t.Errorf("SetupDiSetSelectedDevice(nil) should fail with ERROR_INVALID_USER_BUFFER")
		}
	}
}