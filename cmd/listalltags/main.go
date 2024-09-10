package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/distribution/reference"
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
		tags, err := repository.Tags(ctx).All(ctx)
		if err != nil {
			fmt.Printf("GetTag error:%+v\n", err)
			os.Exit(1)
		}
		for _, tag := range tags {
			fmt.Println(tag)
		}
		return
	}
}
