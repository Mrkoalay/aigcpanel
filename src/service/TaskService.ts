import {TimeUtil} from "../lib/util";
import {TaskChangeType, useTaskStore} from "../store/modules/task";
import {groupThrottle} from "../lib/groupThrottle";
import {t} from "../lang";
import {backend} from "./backend";

const taskStore = useTaskStore();

export type TaskBiz =
    | never
    | "SoundGenerate"
    | "SoundAsr"
    | "VideoGen"
    // sound apps
    | "LongTextTts"
    | "SubtitleTts"
    | "SoundReplace"
    // video apps
    | "TextToImage"
    | "ImageToImage"
    | "VideoGenFlow";

export type TaskJobResultStepStatus = undefined | "queue" | "pending" | "running" | "success" | "fail";

export enum TaskType {
    User = 1,
    System = 2,
}

export type TaskRecord<MODEL_CONFIG extends any = any, JOB_RESULT extends any = any> = {
    id?: number;

    biz: TaskBiz;
    type?: TaskType;

    title: string;

    status?: "queue" | "wait" | "running" | "success" | "fail" | "pause";
    statusMsg?: string;
    startTime?: number;
    endTime?: number | undefined;

    serverName: string;
    serverTitle: string;
    serverVersion: string;

    param?: any;
    jobResult?: JOB_RESULT;
    modelConfig?: MODEL_CONFIG;
    result?: {
        percent: number
    } & any;

    runtime?: TaskRuntime;
};

export type TaskRuntime = {
    [key: string]: any;
};

const mergeData = (oldData: any, newData: any) => {
    if (typeof oldData !== "object" || oldData === null) {
        return newData;
    }
    if (typeof newData !== "object" || newData === null) {
        return newData;
    }
    const result = {};
    for (const key in oldData) {
        if (key in newData) {
            if (Array.isArray(newData[key])) {
                result[key] = newData[key];
            } else if (typeof newData[key] === "object" && newData[key] !== null) {
                result[key] = mergeData(oldData[key], newData[key]);
            } else {
                result[key] = newData[key];
            }
        } else {
            result[key] = oldData[key];
        }
    }
    for (const key in newData) {
        if (!(key in oldData)) {
            result[key] = newData[key];
        }
    }
    return result;
};

type CleanerFunctionType = (record: TaskRecord) => Promise<{
    files: string[];
}>;

type TaskApiRecord = {
    id: number;
    biz: TaskBiz;
    type: number;
    title: string;
    status: string;
    statusMsg: string;
    startTime: number;
    endTime: number;
    serverName: string;
    serverTitle: string;
    serverVersion: string;
    param: string;
    modelConfig: string;
    jobResult: string;
    result: string;
    createdAt: number;
    updatedAt: number;
};

const cleanersMap = new Map<TaskBiz, CleanerFunctionType>();

export const TaskService = {
    registerCleaner(biz: TaskBiz, cleaner: CleanerFunctionType) {
        cleanersMap.set(biz, cleaner);
    },
    decodeRecord(record: TaskApiRecord): TaskRecord | null {
        if (!record) {
            return null;
        }
        return {
            ...record,
            param: JSON.parse(record.param ? record.param : "{}"),
            jobResult: JSON.parse(record.jobResult ? record.jobResult : "{}"),
            modelConfig: JSON.parse(record.modelConfig ? record.modelConfig : "{}"),
            result: JSON.parse(record.result ? record.result : "{}"),
        } as TaskRecord;
    },
    encodeRecord(record: Partial<TaskRecord>): any {
        const result: any = {...record};
        if ("param" in result) {
            result.param = JSON.stringify(result.param || {});
        }
        if ("jobResult" in result) {
            result.jobResult = JSON.stringify(result.jobResult || {});
        }
        if ("modelConfig" in result) {
            result.modelConfig = JSON.stringify(result.modelConfig || {});
        }
        if ("result" in result) {
            result.result = JSON.stringify(result.result || {});
        }
        return result;
    },
    async get(id: number | string): Promise<TaskRecord | null> {
        const record = await backend.get<TaskApiRecord>(`/app/tasks/${id}`);
        return this.decodeRecord(record);
    },
    async list(biz: TaskBiz, type: TaskType = TaskType.User): Promise<TaskRecord[]> {
        const records = await backend.get<TaskApiRecord[]>(`/app/tasks?biz=${encodeURIComponent(biz)}&type=${type}`);
        return records.map(record => this.decodeRecord(record) as TaskRecord).sort((a, b) => (b.id || 0) - (a.id || 0));
    },
    async listByStatus(
        biz: TaskBiz,
        statusList: ("queue" | "wait" | "running" | "success" | "fail")[]
    ): Promise<TaskRecord[]> {
        if (!statusList || statusList.length === 0) {
            return [];
        }
        const records = await backend.get<TaskApiRecord[]>(`/app/tasks?biz=${encodeURIComponent(biz)}&status=${statusList.join(",")}`);
        return records.map(record => this.decodeRecord(record) as TaskRecord).sort((a, b) => (b.id || 0) - (a.id || 0));
    },
    async restoreForTask(biz: TaskBiz) {
        const records = await this.listByStatus(biz, ["running", "wait", "queue"]);
        for (let record of records) {
            await taskStore.dispatch(
                record.biz,
                record.id as any,
                {},
                {
                    status: "queue",
                    runStart: record.startTime,
                    queryInterval: 5 * 1000,
                }
            );
        }
    },
    async submit(record: TaskRecord) {
        record.status = "queue";
        record.startTime = TimeUtil.timestampMS();
        if (!("type" in record)) {
            record.type = TaskType.User;
        }
        const payload = this.encodeRecord(record);
        const created = await backend.post<TaskApiRecord>("/app/tasks", payload);
        await taskStore.dispatch(
            record.biz,
            created.id,
            {},
            {
                queryInterval: 5 * 1000,
            }
        );
        return created.id;
    },
    updatePercent: groupThrottle(async (id: string, percent: number) => {
        const {biz} = await TaskService.update(id, {result: {percent}});
        taskStore.fireChange({
            biz: biz!, bizId: id,
        }, 'change');
    }, 3000, {
        trailing: true,
    }),
    cancelCheck: (biz: TaskBiz, bizId: string) => {
        if (taskStore.shouldCancel(biz, bizId)) {
            throw t('任务已取消');
        }
    },
    updateDelayAndFireChange: groupThrottle(async (
        id: string,
        record: Partial<TaskRecord>,
        fireChangeType: TaskChangeType = "running",
        option?: {
            mergeResult?: boolean;
        }
    ) => {
        const {biz} = await TaskService.update(id, record, option);
        taskStore.fireChange({biz: biz!, bizId: id}, fireChangeType);
    }, 1000, {
        trailing: true,
    }),
    async update(
        id: number | string,
        record: Partial<TaskRecord>,
        option?: {
            mergeResult?: boolean;
        }
    ): Promise<{
        updates: number,
        biz: string | null,
    }> {
        option = Object.assign(
            {
                mergeResult: true,
            },
            option
        );
        let recordOld: TaskRecord | null = null;
        if ("result" in record || "jobResult" in record || "startTime" in record) {
            recordOld = await this.get(id);
            if (option.mergeResult) {
                if ("result" in record) {
                    record.result = mergeData(recordOld?.result, record.result);
                }
                if ("jobResult" in record) {
                    record.jobResult = mergeData(recordOld?.jobResult, record.jobResult);
                }
            }
        }
        await backend.patch(`/app/tasks/${id}`, this.encodeRecord(record));
        return {
            updates: 1,
            biz: recordOld ? recordOld.biz : null,
        }
    },
    async delete(record: TaskRecord) {
        const filesForClean: string[] = [];
        if (record.result) {
            for (const k in record.result) {
                if (record.result[k] && typeof record.result[k] === "string") {
                    if (await window.$mapi.file.isHubFile(record.result[k])) {
                        filesForClean.push(record.result[k]);
                    }
                }
            }
        }
        const cleaner = cleanersMap.get(record.biz);
        if (cleaner) {
            const {files} = await cleaner(record);
            if (files && files.length > 0) {
                filesForClean.push(...files);
            }
        }
        for (const file of filesForClean) {
            await window.$mapi.file.deletes(file);
        }
        await backend.delete(`/app/tasks/${record.id}`);
    },
    async count(
        biz: TaskBiz | null,
        startTime: number = 0,
        endTime: number = 0,
        type: TaskType = TaskType.User
    ): Promise<number> {
        if (!biz) {
            return 0;
        }
        const records = await this.list(biz, type);
        return records.filter(item => {
            const createdAt = (item as any).createdAt || 0;
            if (startTime > 0 && createdAt < startTime) {
                return false;
            }
            return !(endTime > 0 && createdAt > endTime);
        }).length;
    },
};
