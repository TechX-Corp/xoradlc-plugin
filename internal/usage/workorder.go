package usage

import (
	"os"

	"github.com/TechX-Corp/xoradlc-plugin/internal/gitx"
	"github.com/TechX-Corp/xoradlc-plugin/internal/link"
)

// workOrderFor returns the WO code the branch checked out at cwd was mapped to (`xdlc link`), or nil.
// Turns are stamped when parsed, so a later relink or branch switch does not move turns already spent.
func workOrderFor(cwd string) *string {
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	root := gitx.GitRoot(cwd)
	if root == "" {
		return nil
	}
	branch := gitx.GitOut(root, "branch", "--show-current")
	if branch == "" {
		return nil
	}
	repoID, _ := gitx.RepoIdentity(root)
	st, ok := link.LinkOf(repoID, branch)
	if !ok {
		return nil
	}
	wo := st["work_order_id"]
	if wo == "" || wo == "<nil>" {
		return nil
	}
	return &wo
}

func applyWorkOrder(recs []Record, wo *string) {
	for i := range recs {
		if recs[i].WorkOrderID == nil {
			recs[i].WorkOrderID = wo
		}
	}
}
