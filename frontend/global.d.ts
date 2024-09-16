export {}

declare global {
  interface Window {
    Plugins: any
  }

  /**
   * Data format returned by wails events
   */
  type WailsEventsResponse<T = string[]> = {
    name: string
    sender: string
    data: T
    Cancelled: boolean
  }

  /**
   * The variable is initialized in `globalMethods.ts:20`
   */
  var AsyncFunction: FunctionConstructor
}
