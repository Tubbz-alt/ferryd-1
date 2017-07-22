//
// Copyright © 2017 Ikey Doherty <ikey@solus-project.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package libferry

import (
	"encoding/xml"
	"errors"
	"github.com/boltdb/bolt"
	"os"
	"sort"
)

// Index will attempt to write the eopkg index out to disk
// This only requires a read-only database view
func (r *Repository) Index(tx *bolt.Tx, pool *Pool) error {
	var pkgIds []string
	rootBucket := tx.Bucket([]byte(DatabaseBucketRepo)).Bucket([]byte(r.ID)).Bucket([]byte(DatabaseBucketPackage))

	c := rootBucket.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		code := NewGobDecoderLight()
		entry := RepoEntry{}
		if err := code.DecodeType(v, &entry); err != nil {
			return err
		}
		pkgIds = append(pkgIds, entry.Published)
	}

	// Ensure we'll emit in a sane order
	sort.Strings(pkgIds)

	encoder := xml.NewEncoder(os.Stdout)
	encoder.Indent("", "    ")

	// Wrap every output item as Package
	elem := xml.StartElement{
		Name: xml.Name{
			Local: "Package",
		},
	}

	// Ensure we have the start element
	if err := encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: "PISI"}}); err != nil {
		return err
	}

	// TODO: merge distributions.xml here

	for _, pkg := range pkgIds {
		entry, err := pool.GetEntry(tx, pkg)
		if err != nil {
			return err
		}
		if err = encoder.EncodeElement(entry.Meta, elem); err != nil {
			return err
		}
	}

	// TODO: Insert Components, then Groups

	// Now finalise the document
	if err := encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "PISI"}}); err != nil {
		return err
	}

	if err := encoder.Flush(); err != nil {
		return err
	}

	return errors.New("not yet implemented")
}