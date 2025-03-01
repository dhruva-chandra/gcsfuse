package storageutil

import (
	"crypto/md5"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/jacobsa/gcloud/gcs"
	. "github.com/jacobsa/ogletest"
	storagev1 "google.golang.org/api/storage/v1"
)

const TestBucketName string = "gcsfuse-default-bucket"
const TestObjectName string = "gcsfuse/default.txt"

func TestObjectAttrs(t *testing.T) { RunTests(t) }

type objectAttrsTest struct {
}

func init() { RegisterTestSuite(&objectAttrsTest{}) }

func (t objectAttrsTest) TestConvertACLRuleToObjectAccessControlMethod() {
	var attrs = storage.ACLRule{
		Entity:   "allUsers",
		EntityID: "123",
		Role:     "OWNER",
		Domain:   "Domain",
		Email:    "Email",
		ProjectTeam: &storage.ProjectTeam{
			ProjectNumber: "123",
			Team:          "Team",
		},
	}

	objectAccessControl := convertACLRuleToObjectAccessControl(attrs)

	ExpectEq(objectAccessControl.Entity, string(attrs.Entity))
	ExpectEq(objectAccessControl.EntityId, attrs.EntityID)
	ExpectEq(objectAccessControl.Role, string(attrs.Role))
	ExpectEq(objectAccessControl.Domain, attrs.Domain)
	ExpectEq(objectAccessControl.Email, attrs.Email)
	ExpectEq(objectAccessControl.ProjectTeam.ProjectNumber, attrs.ProjectTeam.ProjectNumber)
	ExpectEq(objectAccessControl.ProjectTeam.Team, attrs.ProjectTeam.Team)
}

func (t objectAttrsTest) TestObjectAttrsToBucketObjectMethod() {
	var attrMd5 []byte
	timeAttr := time.Now()
	attrs := storage.ObjectAttrs{
		Bucket:                  TestBucketName,
		Name:                    TestObjectName,
		ContentType:             "ContentType",
		ContentLanguage:         "ContentLanguage",
		CacheControl:            "CacheControl",
		EventBasedHold:          true,
		TemporaryHold:           true,
		RetentionExpirationTime: timeAttr,
		ACL:                     nil,
		PredefinedACL:           "PredefinedACL",
		Owner:                   "Owner",
		Size:                    16,
		ContentEncoding:         "ContentEncoding",
		ContentDisposition:      "ContentDisposition",
		MD5:                     attrMd5,
		CRC32C:                  0,
		MediaLink:               "MediaLink",
		Metadata:                nil,
		Generation:              780,
		Metageneration:          0,
		StorageClass:            "StorageClass",
		Created:                 timeAttr,
		Deleted:                 timeAttr,
		Updated:                 timeAttr,
		CustomerKeySHA256:       "CustomerKeySHA256",
		KMSKeyName:              "KMSKeyName",
		Prefix:                  "Prefix",
		Etag:                    "Etag",
		CustomTime:              timeAttr,
	}
	customeTimeExpected := string(attrs.CustomTime.Format(time.RFC3339))

	var md5Expected [md5.Size]byte
	copy(md5Expected[:], attrs.MD5)

	var acl []*storagev1.ObjectAccessControl
	for _, element := range attrs.ACL {
		acl = append(acl, convertACLRuleToObjectAccessControl(element))
	}

	object := ObjectAttrsToBucketObject(&attrs)

	ExpectEq(object.Name, attrs.Name)
	ExpectEq(object.ContentType, attrs.ContentType)
	ExpectEq(object.ContentLanguage, attrs.ContentLanguage)
	ExpectEq(object.CacheControl, attrs.CacheControl)
	ExpectEq(object.Owner, attrs.Owner)
	ExpectEq(object.Size, attrs.Size)
	ExpectEq(object.ContentEncoding, attrs.ContentEncoding)
	ExpectEq(len(object.MD5), len(&md5Expected))
	ExpectEq(cap(object.MD5), cap(&md5Expected))
	ExpectEq(*object.CRC32C, attrs.CRC32C)
	ExpectEq(object.MediaLink, attrs.MediaLink)
	ExpectEq(object.Metadata, attrs.Metadata)
	ExpectEq(object.Generation, attrs.Generation)
	ExpectEq(object.MetaGeneration, attrs.Metageneration)
	ExpectEq(object.StorageClass, attrs.StorageClass)
	ExpectEq(object.Updated.String(), attrs.Updated.String())
	ExpectEq(object.Deleted.String(), attrs.Deleted.String())
	ExpectEq(object.ContentDisposition, attrs.ContentDisposition)
	ExpectEq(object.CustomTime, customeTimeExpected)
	ExpectEq(object.EventBasedHold, attrs.EventBasedHold)
	ExpectEq(object.Acl, acl)
}

func (t objectAttrsTest) TestConvertObjectAccessControlToACLRuleMethod() {
	objectAccessControl := &storagev1.ObjectAccessControl{
		Entity:   "test_entity",
		EntityId: "test_entity_id",
		Role:     "owner",
		Domain:   "test_domain",
		Email:    "test_email@test.com",
		ProjectTeam: &storagev1.ObjectAccessControlProjectTeam{
			ProjectNumber: "test_project_num",
			Team:          "test_team",
		},
	}

	aclRule := convertObjectAccessControlToACLRule(objectAccessControl)

	ExpectEq(aclRule.Entity, objectAccessControl.Entity)
	ExpectEq(aclRule.EntityID, objectAccessControl.EntityId)
	ExpectEq(aclRule.Role, objectAccessControl.Role)
	ExpectEq(aclRule.Domain, objectAccessControl.Domain)
	ExpectEq(aclRule.Email, objectAccessControl.Email)
	ExpectEq(aclRule.ProjectTeam.ProjectNumber, objectAccessControl.ProjectTeam.ProjectNumber)
	ExpectEq(aclRule.ProjectTeam.Team, objectAccessControl.ProjectTeam.Team)
}

func (t objectAttrsTest) TestSetAttrsInWriterMethod() {
	var crc32c uint32 = 45
	var generationPrecondition int64 = 3
	var metaGenerationPrecondition int64 = 33
	md5Hash := md5.Sum([]byte("testing"))
	timeInRFC3339 := "2006-01-02T15:04:05Z07:00"
	createObjectRequest := gcs.CreateObjectRequest{
		Name:                       "test_object",
		ContentType:                "json",
		ContentEncoding:            "universal",
		CacheControl:               "Medium",
		Metadata:                   map[string]string{"file_name": "test.txt"},
		ContentDisposition:         "Test content disposition",
		CustomTime:                 timeInRFC3339,
		EventBasedHold:             true,
		StorageClass:               "High Accessibility",
		Acl:                        nil,
		Contents:                   strings.NewReader("Creating new object"),
		CRC32C:                     &crc32c,
		MD5:                        &md5Hash,
		GenerationPrecondition:     &generationPrecondition,
		MetaGenerationPrecondition: &metaGenerationPrecondition,
	}
	writer := &storage.Writer{}

	writer = SetAttrsInWriter(writer, &createObjectRequest)

	ExpectEq(writer.Name, createObjectRequest.Name)
	ExpectEq(writer.ContentType, createObjectRequest.ContentType)
	ExpectEq(writer.ContentLanguage, createObjectRequest.ContentLanguage)
	ExpectEq(writer.ContentEncoding, createObjectRequest.ContentEncoding)
	ExpectEq(writer.CacheControl, createObjectRequest.CacheControl)
	ExpectEq(writer.Metadata, createObjectRequest.Metadata)
	ExpectEq(writer.ContentDisposition, createObjectRequest.ContentDisposition)
	parsedTime, _ := time.Parse(time.RFC3339, createObjectRequest.CustomTime)
	ExpectTrue(parsedTime.Equal(writer.CustomTime))
	ExpectEq(writer.EventBasedHold, createObjectRequest.EventBasedHold)
	ExpectEq(writer.StorageClass, createObjectRequest.StorageClass)
	ExpectEq(writer.CRC32C, *createObjectRequest.CRC32C)
	ExpectTrue(writer.SendCRC32C)
	ExpectEq(string(writer.MD5[:]), string(createObjectRequest.MD5[:]))
}
