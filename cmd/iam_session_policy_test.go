// Copyright (c) 2015-2025 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"testing"

	"github.com/minio/pkg/v3/policy"
)

func TestSessionPolicyOwnAccountRequiresAllow(t *testing.T) {
	restrictedPolicy := `{
 "Version": "2012-10-17",
 "Statement": [
  {
   "Effect": "Allow",
   "Action": ["s3:ListBucket"],
   "Resource": ["arn:aws:s3:::bucket1", "arn:aws:s3:::bucket2"]
  },
  {
   "Effect": "Allow",
   "Action": ["s3:GetObject", "s3:PutObject"],
   "Resource": ["arn:aws:s3:::bucket1/*", "arn:aws:s3:::bucket2/*"]
  }
 ]
}`

	args := policy.Args{
		Action:   policy.CreateServiceAccountAdminAction,
		DenyOnly: true,
		IsOwner:  true,
		Claims: map[string]any{
			sessionPolicyNameExtracted: restrictedPolicy,
		},
	}

	hasSP, allowed := isAllowedBySessionPolicy(args)
	if !hasSP {
		t.Fatal("expected STS session policy to be present")
	}
	if allowed {
		t.Fatal("restricted STS session policy must not allow CreateServiceAccount")
	}

	hasSP, allowed = isAllowedBySessionPolicyForServiceAccount(args)
	if !hasSP {
		t.Fatal("expected service-account session policy to be present")
	}
	if allowed {
		t.Fatal("restricted service-account session policy must not allow CreateServiceAccount")
	}

	// Truly unset/empty SA policy still inherits parent (hasSessionPolicy=false).
	args.Claims[sessionPolicyNameExtracted] = "null"
	hasSP, _ = isAllowedBySessionPolicyForServiceAccount(args)
	if hasSP {
		t.Fatal("empty/null service-account policy should inherit parent")
	}
}
