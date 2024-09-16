import i18n from '@/lang'
import { Theme, type MenuItem, Color, Lang } from '@/constant'
import { useAppSettingsStore, useKernelApiStore, useEnvStore, usePluginsStore } from '@/stores'
import { Notify, RestartApp, UpdateTray, UpdateTrayMenus, Window, Events } from '@/bridge'
import {
  debounce,
  exitApp,
  handleChangeMode,
  handleUseProxy,
  sampleID,
  APP_TITLE,
  APP_VERSION
} from '@/utils'

const generateUniqueIdForMenu = (menus: MenuItem[]) => {
  const { t } = i18n.global
  const menuHandlerMap: Record<string, MenuItem['handler']> = {}
  Events.Off('onTrayMenuClick')
  Events.On('onTrayMenuClick', ({ data: [id] }: WailsEventsResponse<[string]>) => {
    menuHandlerMap[id]?.()
  })

  function processMenu(menu: MenuItem) {
    menu.text = t(menu.text || '')
    menu.tooltip = t(menu.tooltip || '')

    if (menu.handler) {
      menu.id = sampleID()
      menuHandlerMap[menu.id] = menu.handler
    }

    if (menu.children && menu.children.length > 0) {
      menu.children = menu.children.map(processMenu)
    }

    return menu
  }

  return menus.map(processMenu)
}

const getTrayMenus = () => {
  const envStore = useEnvStore()
  const appSettings = useAppSettingsStore()
  const kernelApiStore = useKernelApiStore()
  const pluginsStore = usePluginsStore()

  const { providers, proxies } = kernelApiStore

  const groupsMenus: MenuItem[] = (() => {
    if (!providers.default) return []
    return providers.default.proxies
      .concat([proxies.GLOBAL])
      .filter((v) => v.all && !v.hidden)
      .map((group) => {
        const all = group.all
          .filter((proxy) => {
            return (
              appSettings.app.kernel.unAvailable ||
              ['DIRECT', 'REJECT'].includes(proxy) ||
              proxies[proxy].all ||
              proxies[proxy].alive
            )
          })
          .map((proxy) => {
            const history = proxies[proxy].history || []
            const delay = history[history.length - 1]?.delay || 0
            return { ...proxies[proxy], delay }
          })
          .sort((a, b) => {
            if (!appSettings.app.kernel.sortByDelay || a.delay === b.delay) return 0
            if (!a.delay) return 1
            if (!b.delay) return -1
            return a.delay - b.delay
          })
        return { ...group, all }
      })
      .map((group) => {
        return {
          type: 'item',
          text: group.name,
          children: group.all.map((proxy) => {
            return {
              type: 'radio',
              text: proxy.name,
              show: true,
              checked: proxy.name === group.now,
              handler: () => {
                handleUseProxy(group, proxy)
              }
            }
          })
        }
      })
  })()

  let pluginMenus: MenuItem[] = []
  let pluginMenusHidden = !appSettings.app.addPluginToMenu

  if (!pluginMenusHidden) {
    const filtered = pluginsStore.plugins.filter(
      (plugin) => Object.keys(plugin.menus).length && !plugin.disabled
    )
    pluginMenusHidden = filtered.length === 0
    pluginMenus = filtered.map(({ id, name, menus }) => {
      return {
        type: 'item',
        text: name,
        children: Object.entries(menus).map(([text, event]) => {
          return {
            type: 'item',
            text,
            event: () => {
              pluginsStore.manualTrigger(id, event as any).catch((err: any) => {
                Notify('Error', err.message || err)
              })
            }
          }
        })
      }
    })
  }

  const trayMenus: MenuItem[] = [
    {
      type: 'item',
      text: 'kernel.mode',
      hidden: !appSettings.app.kernel.running,
      children: [
        {
          type: 'radio',
          text: 'kernel.global',
          checked: kernelApiStore.config.mode === 'global',
          handler: () => handleChangeMode('global')
        },
        {
          type: 'radio',
          text: 'kernel.rule',
          checked: kernelApiStore.config.mode === 'rule',
          handler: () => handleChangeMode('rule')
        },
        {
          type: 'radio',
          text: 'kernel.direct',
          checked: kernelApiStore.config.mode === 'direct',
          handler: () => handleChangeMode('direct')
        }
      ]
    },
    {
      type: 'item',
      text: 'tray.proxyGroup',
      hidden: !appSettings.app.kernel.running,
      children: groupsMenus
    },
    {
      type: 'item',
      text: 'tray.kernel',
      children: [
        {
          type: 'item',
          text: 'tray.startKernel',
          hidden: appSettings.app.kernel.running,
          handler: kernelApiStore.startKernel
        },
        {
          type: 'item',
          text: 'tray.restartKernel',
          hidden: !appSettings.app.kernel.running,
          handler: kernelApiStore.restartKernel
        },
        {
          type: 'item',
          text: 'tray.stopKernel',
          hidden: !appSettings.app.kernel.running,
          handler: kernelApiStore.stopKernel
        }
      ]
    },
    {
      type: 'separator',
      hidden: !appSettings.app.kernel.running
    },
    {
      type: 'item',
      text: 'tray.proxy',
      hidden: !appSettings.app.kernel.running,
      children: [
        {
          type: 'item',
          text: 'tray.setSystemProxy',
          hidden: envStore.systemProxy,
          handler: async () => {
            await kernelApiStore.updateConfig({ tun: { enable: false } })
            await envStore.setSystemProxy()
          }
        },
        {
          type: 'item',
          text: 'tray.clearSystemProxy',
          hidden: !envStore.systemProxy,
          handler: envStore.clearSystemProxy
        }
      ]
    },
    {
      type: 'item',
      text: 'tray.tun',
      hidden: !appSettings.app.kernel.running,
      children: [
        {
          type: 'item',
          text: 'tray.enableTunMode',
          hidden: kernelApiStore.config.tun.enable,
          handler: async () => {
            await envStore.clearSystemProxy()
            await kernelApiStore.updateConfig({ tun: { enable: true } })
          }
        },
        {
          type: 'item',
          text: 'tray.disableTunMode',
          hidden: !kernelApiStore.config.tun.enable,
          handler: async () => {
            await kernelApiStore.updateConfig({ tun: { enable: false } })
          }
        }
      ]
    },
    {
      type: 'item',
      text: 'settings.general',
      children: [
        {
          type: 'item',
          text: 'settings.theme.name',
          children: [
            {
              type: 'radio',
              text: 'settings.theme.dark',
              checked: appSettings.app.theme === Theme.Dark,
              handler: () => (appSettings.app.theme = Theme.Dark)
            },
            {
              type: 'radio',
              text: 'settings.theme.light',
              checked: appSettings.app.theme === Theme.Light,
              handler: () => (appSettings.app.theme = Theme.Light)
            },
            {
              type: 'radio',
              text: 'settings.theme.auto',
              checked: appSettings.app.theme === Theme.Auto,
              handler: () => (appSettings.app.theme = Theme.Auto)
            }
          ]
        },
        {
          type: 'item',
          text: 'settings.color.name',
          children: [
            {
              type: 'radio',
              text: 'settings.color.default',
              checked: appSettings.app.color === Color.Default,
              handler: () => (appSettings.app.color = Color.Default)
            },
            {
              type: 'radio',
              text: 'settings.color.orange',
              checked: appSettings.app.color === Color.Orange,
              handler: () => (appSettings.app.color = Color.Orange)
            },
            {
              type: 'radio',
              text: 'settings.color.pink',
              checked: appSettings.app.color === Color.Pink,
              handler: () => (appSettings.app.color = Color.Pink)
            },
            {
              type: 'radio',
              text: 'settings.color.red',
              checked: appSettings.app.color === Color.Red,
              handler: () => (appSettings.app.color = Color.Red)
            },
            {
              type: 'radio',
              text: 'settings.color.skyblue',
              checked: appSettings.app.color === Color.Skyblue,
              handler: () => (appSettings.app.color = Color.Skyblue)
            },
            {
              type: 'radio',
              text: 'settings.color.green',
              checked: appSettings.app.color === Color.Green,
              handler: () => (appSettings.app.color = Color.Green)
            }
          ]
        },
        {
          type: 'item',
          text: 'settings.lang.name',
          children: [
            {
              type: 'radio',
              text: 'settings.lang.zh',
              checked: appSettings.app.lang === Lang.ZH,
              handler: () => (appSettings.app.lang = Lang.ZH)
            },
            {
              type: 'radio',
              text: 'settings.lang.en',
              checked: appSettings.app.lang === Lang.EN,
              handler: () => (appSettings.app.lang = Lang.EN)
            }
          ]
        }
      ]
    },
    {
      type: 'item',
      text: 'tray.plugins',
      hidden: pluginMenusHidden,
      children: pluginMenus
    },
    {
      type: 'separator'
    },
    {
      type: 'item',
      text: 'tray.restart',
      tooltip: 'tray.restartTip',
      handler: RestartApp
    },
    {
      type: 'item',
      text: 'tray.exit',
      tooltip: 'tray.exitTip',
      handler: exitApp
    }
  ]

  return generateUniqueIdForMenu(trayMenus)
}

const getTrayIcons = () => {
  const envStore = useEnvStore()
  const appSettings = useAppSettingsStore()
  const kernelApiStore = useKernelApiStore()

  const themeMode = appSettings.themeMode
  let icon = `data/.cache/icons/tray_normal_${themeMode}.png`

  if (appSettings.app.kernel.running) {
    if (kernelApiStore.config.tun.enable) {
      icon = `data/.cache/icons/tray_tun_${themeMode}.png`
    } else if (envStore.systemProxy) {
      icon = `data/.cache/icons/tray_proxy_${themeMode}.png`
    }
  }
  return icon
}

export const updateTrayMenus = debounce(async () => {
  const trayMenus = getTrayMenus()
  const trayIcons = getTrayIcons()
  await UpdateTray({ icon: trayIcons, title: APP_TITLE, tooltip: APP_TITLE + ' ' + APP_VERSION })
  await UpdateTrayMenus(trayMenus as any)
}, 500)
