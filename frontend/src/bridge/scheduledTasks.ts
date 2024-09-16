import { App } from '@wails/guiforcores/bridge'

export const AddScheduledTask = async (cron: string, event: string) => {
  const { flag, data } = await App.AddScheduledTask(cron, event)
  if (!flag) {
    throw data
  }
  return Number(data)
}

export const RemoveScheduledTask = async (id: number) => {
  await App.RemoveScheduledTask(id)
}

export const ValidateCron = async (cron: string) => {
  const { flag, data } = await App.ValidateCron(cron)
  if (!flag) {
    throw data
  }
  return data
}
