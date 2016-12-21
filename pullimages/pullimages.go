package pullimages

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/verbose"
)

func getLocallyAvailableImagesList() (map[string]bool, error) {
	var images = map[string]bool{}
	var params = []string{
		"images",
		"--format",
		"{{.Repository}}:{{.Tag}}",
	}

	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var list = exec.Command("docker", params...)
	var buf bytes.Buffer
	list.Stderr = os.Stderr
	list.Stdout = &buf

	if err := list.Run(); err != nil {
		return images, err
	}

	// clear leading line break
	var out = strings.TrimSuffix(buf.String(), "\n")

	for _, id := range strings.Split(out, "\n") {
		if !images[id] {
			images[id] = true
		}
	}

	return images, nil
}

func getContainerTypesFromContainersDirectories(csDirs []string) (containersImages map[string]string, err error) {
	containersImages = map[string]string{}
	for _, c := range csDirs {
		container, err := containers.Read(c)

		if err != nil {
			return nil, errwrap.Wrapf("Failure trying to read containers types availability: {{err}}", err)
		}

		containersImages[c] = container.Type
	}

	return containersImages, nil
}

func pullImage(image string) (err error) {
	var pullMetrics = &pullMetrics{
		image: image,
	}

	fmt.Printf(color.Format(color.FgHiBlue, "Pulling image %v\n", image))
	pullMetrics.reportStart()
	var docker = exec.Command("docker", "pull", image)
	docker.Stderr = os.Stderr
	docker.Stdout = os.Stdout

	err = docker.Run()
	fmt.Println("")

	switch err {
	case nil:
		pullMetrics.reportSuccess()
		return nil
	default:
		pullMetrics.reportError()
		err = errwrap.Wrapf("Image pull error: {{err}}", err)
	}

	return err
}

func getMissingContainersTypes(typesFromContainers map[string]string, locallyAvailable map[string]bool) (missing []string) {
	var inMissingList = map[string]bool{}

	for c, i := range typesFromContainers {
		if !strings.Contains(i, ":") {
			i = i + ":latest"
		}

		if locallyAvailable[i] {
			continue
		}

		verbose.Debug(fmt.Sprintf("Container %v requires missing image %v", c, i))

		if !inMissingList[i] {
			missing = append(missing, i)
			inMissingList[i] = true
		}
	}

	return missing
}

// PullMissingContainersImages pulls missing images using docker pull on the foreground
func PullMissingContainersImages(csDirs []string) (err error) {
	var locallyAvailable, errAvailable = getLocallyAvailableImagesList()

	if errAvailable != nil {
		return errwrap.Wrapf("Error trying to list locally available images: {{err}}", errAvailable)
	}

	var typesFromContainers, errGetTypes = getContainerTypesFromContainersDirectories(csDirs)

	if errGetTypes != nil {
		return errGetTypes
	}

	var missing = getMissingContainersTypes(typesFromContainers, locallyAvailable)

	if len(missing) == 0 {
		return nil
	}

	fmt.Println("Pulling required missing docker containers:")

	for _, needed := range missing {
		fmt.Printf("\t%v\n", strings.TrimSuffix(needed, ":latest"))
	}

	fmt.Println("")
	return pullImages(missing)
}

func pullImages(missing []string) (err error) {
	for _, needed := range missing {
		// currently we don't want to download in parallel
		// and use this wait group mostly for a side-effect
		// (intercepting signals to the main thread)
		// instead of handling signals on our own
		var queue sync.WaitGroup
		queue.Add(1)

		go func() {
			if err = pullImage(needed); err != nil {
				err = errwrap.Wrapf("Error while trying to pull image", err)
			}

			queue.Done()
		}()

		queue.Wait()

		if err != nil {
			return err
		}
	}

	if len(missing) != 0 {
		fmt.Println(color.Format(color.FgHiGreen, "Number of container images pulled: %v\n", len(missing)))
	}

	return nil
}
