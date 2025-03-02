import {ak_call as _ak_call} from "./ak/ak_call";
import {mockWaiter} from "./MockWaiter";
global.ak_call = _ak_call(new mockWaiter());
