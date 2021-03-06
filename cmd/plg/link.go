package main

import (
	"errors"
	"fmt"

	"github.com/gbrlsnchs/cli"
	"github.com/gbrlsnchs/pilgo/config"
	"github.com/gbrlsnchs/pilgo/fs"
	"github.com/gbrlsnchs/pilgo/linker"
	"github.com/gbrlsnchs/pilgo/parser"
	"gopkg.in/yaml.v3"
)

type linkCmd struct{}

func (*linkCmd) register(getcfg func() appConfig) func(cli.Program) error {
	return func(prg cli.Program) error {
		appcfg := getcfg()
		exe := prg.Name()
		fs := fs.New(appcfg.fs)
		cwd, err := appcfg.getwd()
		if err != nil {
			return err
		}
		b, err := fs.ReadFile(appcfg.conf)
		if err != nil {
			return err
		}
		var c config.Config
		if yaml.Unmarshal(b, &c); err != nil {
			return err
		}
		baseDir, err := appcfg.userConfigDir()
		if err != nil {
			return err
		}
		var p parser.Parser
		tr, err := p.Parse(c, parser.BaseDir(baseDir), parser.Cwd(cwd), parser.Envsubst)
		if err != nil {
			return err
		}
		ln := linker.New(fs)
		if err := ln.Link(tr); err != nil {
			var cft *linker.ConflictError
			if errors.As(err, &cft) {
				errw := prg.Stderr()
				for _, err := range cft.Errs {
					fmt.Fprintf(errw, "%s: %v\n", exe, err)
				}
			}
			return err
		}
		return nil
	}
}
