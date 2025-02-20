//
// Copyright © 2024 Agora
// This file is part of TEN Framework, an open source project.
// Licensed under the Apache License, Version 2.0, with certain conditions.
// Refer to the "LICENSE" file in the root directory for more information.
//

package default_extension_go

import (
	"fmt"
	"sync/atomic"

	"ten_framework/ten"
)

const concurrency = 10000

type extensionB struct {
	ten.DefaultExtension
}

func newExtensionB(name string) ten.Extension {
	return &extensionB{}
}

func (p *extensionB) OnCmd(
	tenEnv ten.TenEnv,
	cmd ten.Cmd,
) {
	go func() {
		fmt.Println("extensionB OnCmd")

		connected, err := tenEnv.IsCmdConnected("cmd_not_exist")
		if err != nil || connected {
			panic("Should not happen.")
		}

		cmdName, _ := cmd.GetName()
		if cmdName == "B" {
			var count uint32 = 0

			done := make(chan struct{}, 1)
			defer close(done)

			for i := 0; i < concurrency; i++ {
				go func(i int) {
					res, err := cmd.GetPropertyInt64(
						fmt.Sprintf("prop_%d", i),
					)
					if err != nil || int(res) != i {
						panic("should not happen")
					}

					cmdName, _ := cmd.GetName()
					if cmdName != "B" {
						panic("should not happen")
					}

					if atomic.AddUint32(&count, 1) == concurrency {
						done <- struct{}{}
					}
				}(i % 100)
			}
			<-done

			statusCmd, err := ten.NewCmdResult(
				ten.StatusCodeOk,
			)
			if err != nil {
				cmdResult, _ := ten.NewCmdResult(ten.StatusCodeError)
				cmdResult.SetPropertyString("detail", err.Error())
				tenEnv.ReturnResult(cmdResult, cmd)
				return
			}

			statusCmd.SetProperty("detail", "this is extensionB.")
			statusCmd.SetProperty("password", "password")
			tenEnv.ReturnResult(statusCmd, cmd)
		} else {
			cmdResult, _ := ten.NewCmdResult(ten.StatusCodeError)
			cmdResult.SetPropertyString("detail", "wrong cmd name")
			tenEnv.ReturnResult(cmdResult, cmd)
		}
	}()
}

func init() {
	// Register addon
	err := ten.RegisterAddonAsExtension(
		"extension_b",
		ten.NewDefaultExtensionAddon(newExtensionB),
	)
	if err != nil {
		fmt.Println("register addon failed", err)
	}
}
