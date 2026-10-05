package core

import (
	"context"
	"testing"

	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
)

func TestPersisterClonesMessages(t *testing.T) {
	persister := &dataPersisterImpl{}
	for _, operation := range []string{"create", "update"} {
		t.Run(operation, func(t *testing.T) {
			id, name := uint64(7), "original"
			input := &cmdb.CisInfo{Id: &id, CreatedBy: &name}
			var output *cmdb.CisInfo
			var err error
			if operation == "create" {
				output, err = persister.Create(context.Background(), input)
			} else {
				output, err = persister.Update(context.Background(), 8, input)
			}
			if err != nil {
				t.Fatal(err)
			}
			if output == input || output.GetId() == input.GetId() || output.GetCreatedBy() != input.GetCreatedBy() {
				t.Fatal("operation did not preserve payload and assign a new ID")
			}
			*output.CreatedBy = "changed"
			if input.GetCreatedBy() != "original" || input.GetId() != 7 {
				t.Fatal("output mutation affected input")
			}
		})
	}
}
