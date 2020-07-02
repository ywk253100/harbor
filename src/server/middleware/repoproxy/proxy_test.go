//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package repoproxy

import "testing"

func TestValidProxyPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{
			name: `normal`,
			in:   "dockerhub_proxy/library/hello-world",
			want: true,
		},
		{
			name: `invalid`,
			in:   "dockerhub_proxy/hello-world",
			want: false,
		},
		{
			name: `nested`,
			in:   "dockerhub_proxy/library/example/hello-world",
			want: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			got := validProxyRepo(tt.in)

			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}

		})
	}
}
