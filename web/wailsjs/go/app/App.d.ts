// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {config} from '../models';

export function EnableConfig(arg1:string,arg2:string):Promise<void>;

export function GetConfig(arg1:string,arg2:string):Promise<config.Config>;

export function ListConfigs():Promise<{[key: string]: boolean}>;

export function RemoveConfig(arg1:string,arg2:boolean):Promise<void>;

export function SetConfig(arg1:config.Config,arg2:string,arg3:string):Promise<void>;
