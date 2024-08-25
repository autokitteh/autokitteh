package bidirunmsgs

import (
	"fmt"

	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
)

func DescribeReq(req *runtimesv1.BidiRunRequest) string {
	switch req := req.Request.(type) {
	case *runtimesv1.BidiRunRequest_Start_:
		return "start"
	case *runtimesv1.BidiRunRequest_Start1_:
		return "start1"
	case *runtimesv1.BidiRunRequest_Call_:
		return fmt.Sprintf("call %s", req.Call.Value.GetFunction().GetName())
	case *runtimesv1.BidiRunRequest_CallReturn:
		switch req.CallReturn.Result.(type) {
		case *runtimesv1.BidiRunCallReturn_Error:
			return "call return error"
		case *runtimesv1.BidiRunCallReturn_Value:
			return "call return value"
		default:
			return "call return - unknown type"
		}
	case *runtimesv1.BidiRunRequest_LoadReturn:
		if req.LoadReturn.GetError() != nil {
			return "load return error"
		} else if req.LoadReturn.GetValues() != nil {
			return "load return values"
		} else {
			return "load return - no values"
		}
	case *runtimesv1.BidiRunRequest_NewRunIdValue:
		return "new run id value"
	default:
		return "unknown"
	}
}

func DescribeRes(res *runtimesv1.BidiRunResponse) string {
	switch res := res.Response.(type) {
	case *runtimesv1.BidiRunResponse_StartReturn:
		if res.StartReturn.GetError() != nil {
			return "start return error"
		} else if res.StartReturn.GetValues() != nil {
			return "start return values"
		} else {
			return "start return - no values"
		}
	case *runtimesv1.BidiRunResponse_CallReturn:
		switch res.CallReturn.Result.(type) {
		case *runtimesv1.BidiRunCallReturn_Error:
			return "call return error"
		case *runtimesv1.BidiRunCallReturn_Value:
			return "call return value"
		default:
			return "call return - unknown type"
		}
	case *runtimesv1.BidiRunResponse_Print_:
		return "print"
	case *runtimesv1.BidiRunResponse_Call:
		return fmt.Sprintf("call %s", res.Call.Value.GetFunction().GetName())
	case *runtimesv1.BidiRunResponse_Load_:
		return fmt.Sprintf("load: %q", res.Load.GetPath())
	case *runtimesv1.BidiRunResponse_NewRunId:
		return "new run id"
	default:
		return "unknown"
	}
}
