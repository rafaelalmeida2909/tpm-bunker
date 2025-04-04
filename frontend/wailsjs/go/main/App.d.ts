// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {types} from '../models';

export function AuthLogin():Promise<boolean>;

export function CheckConnection():Promise<boolean>;

export function CheckTPMPresence():Promise<boolean>;

export function DecryptFile(arg1:string):Promise<void>;

export function EncryptFile(arg1:string):Promise<void>;

export function GetDeviceInfo():Promise<types.DeviceInfo>;

export function GetOperations():Promise<Array<number>>;

export function GetTPMStatus():Promise<types.TPMStatus>;

export function InitializeDevice():Promise<types.DeviceInfo>;

export function IsDeviceInitialized():Promise<boolean>;

export function SelectFile():Promise<string>;
