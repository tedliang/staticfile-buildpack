package cutlass

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

func executeDockerFile(fixture_path, buildpack_path, network_command string) error {

  dockerfile_path = "Dockerfile.#{$PROCESS_ID}.#{Time.now.to_i}"
  docker_image_name = 'internet_traffic_test'

  // docker_env_vars += get_app_env_vars(fixture_path)

  dockerfile_contents = dockerfile(docker_env_vars, fixture_path, buildpack_path, network_command)

  File.write(dockerfile_path, dockerfile_contents)

  exit_status, output = execute_test_in_docker_container(dockerfile_path, docker_image_name)
  [exit_status, output, dockerfile_path]
}

func dockerfile(fixture_path, buildpack_path, network_command string) string {
	out := "FROM cloudfoundry/cflinuxfs2\n" +
    out += "ENV CF_STACK cflinuxfs2\n" +
    out += "ENV VCAP_APPLICATION {}\n" +
	// TODO env vars
    // out += "#{env_vars}\n" +
    out += "ADD "+fixture_path+" /tmp/staged/\n" +
    out += "ADD ./"+buildpack_path+" /tmp/\n" +
    out += "RUN mkdir -p /buildpack\n" +
    out += "RUN mkdir -p /tmp/cache\n" +
    out += "RUN unzip /tmp/"+buildpack_path+" -d /buildpack\n" +
    out += "# HACK around https://github.com/dotcloud/docker/issues/5490\n" +
    out += "RUN mv /usr/sbin/tcpdump /usr/bin/tcpdump\n" +
    out += "RUN "+network_command+"\n"
	return out
}
