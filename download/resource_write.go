package download

import (
	"errors"
	"os"
	"strings"

	"github.com/dtcookie/hcl"
	"github.com/dynatrace-oss/terraform-provider-dynatrace/hclgen"
)

func (me ResourceData) WriteResourceSeparate(dlConfig DownloadConfig, resName string, resFolder string, resources Resources, resNameCnt NameCounter) error {
	var err error
	for _, resource := range resources {
		if resource.ReqInter {
			continue
		}

		var file *os.File
		fileName := dlConfig.TargetFolder + "/" + resFolder + "/" + resFolder + "." + escf(resource.Name) + ".tf"
		os.Remove(fileName)
		if file, err = os.Create(fileName); err != nil {
			return err
		}

		if dlConfig.CommentedID {
			if err := hclgen.Export(resource.RESTObject, file, resName, resNameCnt.Numbering(escape(resource.Name)), "id = "+resource.ID); err != nil {
				file.Close()
				return err
			}
		} else {
			if err := hclgen.Export(resource.RESTObject, file, resName, escape(resource.Name)); err != nil {
				file.Close()
				return err
			}
		}

		if resName == "dynatrace_dashboard" {
			if err := me.writeDashboardSharing(file, resource.Name); err != nil {
				file.Close()
				return err
			}
		}
		file.Close()
	}

	return nil
}

func (me ResourceData) writeDashboardSharing(file *os.File, name string) error {
	var restObject hcl.Marshaler
	var found bool
	for _, resource := range me["dynatrace_dashboard_sharing"] {
		if resource.Name == name {
			restObject = resource.RESTObject
			found = true
			break
		}
	}
	if !found {
		file.Close()
		return nil
	}
	if err := hclgen.Export(restObject, file, "dynatrace_dashboard_sharing", escape(name)); err != nil {
		file.Close()
		return err
	}
	return nil
}

func (me ResourceData) WriteResReqAttn(dlConfig DownloadConfig) error {
	var err error
	for resName := range InterventionInfoMap {
		if resources, exists := me[resName]; exists {
			for _, resource := range resources {
				if !resource.ReqInter {
					continue
				}

				folderName := dlConfig.TargetFolder + "/" + ".requires_attention"
				if _, err := os.Stat(folderName); errors.Is(err, os.ErrNotExist) {
					err := os.Mkdir(folderName, os.ModePerm)
					if err != nil {
						return err
					}
				}

				var file *os.File
				fileName := folderName + "/" + strings.TrimPrefix(resName, "dynatrace_") + "." + escf(resource.Name) + ".tf"
				os.Remove(fileName)
				if file, err = os.Create(fileName); err != nil {
					return err
				}

				if dlConfig.CommentedID {
					if err := hclgen.Export(resource.RESTObject, file, resName, escape(resource.Name), "id = "+resource.ID); err != nil {
						file.Close()
						return err
					}
				} else {
					if err := hclgen.Export(resource.RESTObject, file, resName, escape(resource.Name)); err != nil {
						file.Close()
						return err
					}
				}

				file.Close()
			}
		}
	}
	return nil
}
