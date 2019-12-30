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

package testing

import (
	"github.com/stretchr/testify/mock"
)

// FakeBlobReader is a fake blob reader that implement the src/api/artifact/blob.Reader interface
type FakeBlobReader struct {
	mock.Mock
}

// Read ...
func (f *FakeBlobReader) Read(name, digest string, manifest bool) (string, []byte, error) {
	args := f.Called(mock.Anything)
	return args.String(0), args.Get(1).([]byte), args.Error(2)
}
