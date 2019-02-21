// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chart

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/replication/ng/model"
	trans "github.com/goharbor/harbor/src/replication/ng/transfer"
)

var (
	jobStoppedErr = errs.JobStoppedError()
)

func init() {
	if err := trans.RegisterFactory(model.ResourceTypeChart, factory); err != nil {
		log.Errorf("failed to register transfer factory: %v", err)
	}
}

func factory(logger trans.Logger, stopFunc trans.CancelFunc) (trans.Transfer, error) {
	return &transfer{
		logger:    logger,
		isStopped: stopFunc,
	}, nil
}

type transfer struct {
	logger    trans.Logger
	isStopped trans.CancelFunc
	src       Registry
	dst       Registry
}

func (t *transfer) Transfer(src *model.Resource, dst *model.Resource) error {
	// initialize
	if err := t.initialize(src.Registry, dst.Registry); err != nil {
		return err
	}

	// delete the chart on destination registry
	if dst.Deleted {
		return t.delete(dst)
	}

	// copy the chart from source registry to the destination
	return t.copy(src, dst)
}

func (t *transfer) shouldStop() bool {
	isStopped := t.isStopped()
	if isStopped {
		t.logger.Info("the job is stopped")
	}
	return isStopped
}

func (t *transfer) initialize(src, dst *model.Registry) error {
	if t.shouldStop() {
		return jobStoppedErr
	}
	t.src = NewRegistry(src)
	t.logger.Infof("client for source registry [type: %s, URL: %s, insecure: %v] created",
		src.Type, src.URL, src.Insecure)
	t.dst = NewRegistry(dst)
	t.logger.Infof("client for destination registry [type: %s, URL: %s, insecure: %v] created",
		dst.Type, dst.URL, dst.Insecure)
	return nil
}

func (t *transfer) copy(src, dst *model.Resource) error {
	if t.shouldStop() {
		return jobStoppedErr
	}
	t.logger.Infof("copying %s(source registry) to %s(destination registry)...", src.URI, dst.URI)

	path := dst.URI + "/" + dst.Metadata.Name
	exist, err := t.dst.Exist(path)
	if err != nil {
		t.logger.Errorf("failed to check the existence of %s on the destination registry: %v", path, err)
		return err
	}

	if exist {
		if !dst.Override {
			t.logger.Warningf("the same name chart %s exists on the destination registry, but the \"override\" is set to false, skip",
				path)
			return nil
		}
		t.logger.Warningf("the same name chart %s exists on the destination registry and the \"override\" is set to true, continue...",
			path)
	}

	chart, err := t.src.Download(src.URI)
	if err != nil {
		t.logger.Errorf("failed to download the chart: %v", err)
		return err
	}

	if t.shouldStop() {
		return jobStoppedErr
	}

	err = t.dst.Upload(dst.URI, chart)
	if err != nil {
		t.logger.Errorf("failed to upload the chart: %v", err)
		return err
	}

	t.logger.Infof("copy %s(source registry) to %s(destination registry) completed", src.URI, dst.URI)
	return nil
}

func (t *transfer) delete(res *model.Resource) error {
	if t.shouldStop() {
		return jobStoppedErr
	}

	t.logger.Infof("deleting %s on destination registry...", res.URI)
	err := t.dst.Delete(res.URI)
	if err != nil {
		t.logger.Errorf("failed to delete the chart: %v", err)
		return err
	}
	t.logger.Infof("%s deleted", res.URI)
	return nil
}
