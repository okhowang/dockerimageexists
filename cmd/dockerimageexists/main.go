package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/api/errcode"
	v2 "github.com/docker/distribution/registry/api/v2"
	apiregistry "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/distribution"
	"github.com/docker/docker/registry"
)

func main() {
	image := flag.String("image", "", "docker image")
	username := flag.String("username", "", "username")
	usernameFile := flag.String("username_file", "", "username from file")
	password := flag.String("password", "", "password")
	passwordFile := flag.String("password_file", "", "password from file")
	outFile := flag.String("outfile", "", "output to file if found")
	exitCode := flag.Int("exit_code", 0, "exit code when image not found")
	flag.Parse()

	flag.VisitAll(func(f *flag.Flag) {
		env := "PLUGIN_" + strings.ToUpper(f.Name)
		if v, ok := os.LookupEnv(env); ok {
			f.Value.Set(v)
		}
	})

	if *image == "" {
		fmt.Printf("Image can not be empty\n")
		os.Exit(1)
	}
	named, err := reference.ParseNormalizedNamed(*image)
	if err != nil {
		fmt.Printf("ParseNormalizedNamed error:%s\n", err)
		os.Exit(1)
	}
	tagged, ok := named.(reference.Tagged)
	if !ok {
		fmt.Printf("Image must have tag\n")
		os.Exit(1)
	}

	ctx := context.Background()
	registryService, err := registry.NewService(registry.ServiceOptions{})
	if err != nil {
		fmt.Printf("NewService error:%s\n", err)
		os.Exit(1)
	}

	if *usernameFile != "" {
		data, err := ioutil.ReadFile(*usernameFile)
		if err != nil {
			fmt.Printf("readfile %s error:%s\n", *usernameFile, err)
			os.Exit(1)
		}
		*username = string(data)
	}
	if *passwordFile != "" {
		data, err := ioutil.ReadFile(*passwordFile)
		if err != nil {
			fmt.Printf("readfile %s error:%s\n", *passwordFile, err)
			os.Exit(1)
		}
		*password = string(data)
	}
	autoConfig := &apiregistry.AuthConfig{
		Username: *username,
		Password: *password,
	}
	repositories, err := distribution.GetRepositories(ctx, named, &distribution.ImagePullConfig{
		Config: distribution.Config{
			RegistryService: registryService,
			AuthConfig:      autoConfig,
		},
	})
	if err != nil {
		fmt.Printf("NewV2Repository error:%s\n", err)
		os.Exit(1)
	}
	for _, repository := range repositories {
		tag, err := repository.Tags(ctx).Get(ctx, tagged.Tag())
		if err != nil {
			var errs errcode.Errors
			var apiErr errcode.Error
			if errors.As(err, &errs) && errs.Len() == 1 && errors.As(errs[0], &apiErr) &&
				errors.Is(apiErr.Code, v2.ErrorCodeManifestUnknown) {
				fmt.Printf("Image not found\n")
				os.Exit(*exitCode)
			}
			fmt.Printf("GetTag error:%+v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%+v\n", tag)
		if *outFile != "" {
			err = ioutil.WriteFile(*outFile, []byte{}, 0644)
			if err != nil {
				fmt.Printf("WriteFile err:%s\n", err)
				os.Exit(1)
			}
		}
		return
	}
}
