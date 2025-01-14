// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {types} from '../models';
import {tpm} from '../models';

export function ExecuteOperation(arg1:types.UserOperation):Promise<types.APIResponse>;

export function GetDeviceInfo():Promise<types.DeviceInfo>;

export function GetTPMStatus():Promise<tpm.TPMStatus>;

export function InitializeDevice():Promise<types.DeviceInfo>;

export function IsDeviceInitialized():Promise<boolean>;
