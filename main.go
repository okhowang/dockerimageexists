package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/distribution"
	"github.com/docker/docker/registry"
	"io/ioutil"
	"os"
	"strings"
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

	endpoints, err := registryService.LookupPullEndpoints(reference.Domain(named))
	if err != nil {
		fmt.Printf("LookupPullEndpoints error:%s\n", err)
		os.Exit(1)
	}

	repoInfo, err := registryService.ResolveRepository(named)
	if err != nil {
		fmt.Printf("ResolveRepository error:%s\n", err)
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
	autoConfig := &types.AuthConfig{
		Username: *username,
		Password: *password,
	}
	for _, endpoint := range endpoints {
		if endpoint.Version == registry.APIVersion1 {
			continue
		}
		repository, _, err := distribution.NewV2Repository(ctx, repoInfo, endpoint, nil, autoConfig, "pull")
		if err != nil {
			fmt.Printf("NewV2Repository error:%s\n", err)
			continue
		}
		tag, err := repository.Tags(ctx).Get(ctx, tagged.Tag())
		if err != nil {
			fmt.Printf("GetTag error:%+v\n", err)
			continue
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
	fmt.Printf("No more endpoint\n")
	os.Exit(*exitCode)
}
