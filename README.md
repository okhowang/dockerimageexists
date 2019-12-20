# dockerimageexists
detect docker image exists on remote without pull.

## Usage

```shell
dockerimageexists -image ubuntu:18.04
# need auth
dockerimageexists -image some_private:tag -username $user -password $pwd
# output file (empty file currently) when found
dockerimageexists -image some_private:tag -outfile $filename
```

## CI Plugin

```yaml
          image: wangpeiwen/dockerimageexists
          settings:
            username: someone
            password: pwd
            image: ubuntu:18.04
            outfile: outfile
```

