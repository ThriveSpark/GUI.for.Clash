<script setup lang="ts">
import { computed, inject, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import * as Defaults from '@/constant'
import { DefaultSubscribeScript } from '@/constant/app'
import {
  useProfilesStore,
  useAppSettingsStore,
  useSubscribesStore,
  type ProfileType,
} from '@/stores'
import { message, sampleID } from '@/utils'

import type { Subscription } from '@/types/app'

const url = ref('')
const loading = ref(false)

const { t } = useI18n()
const subscribeStore = useSubscribesStore()
const profilesStore = useProfilesStore()
const appSettingsStore = useAppSettingsStore()

const handleCancel = inject('cancel') as any

const canSubmit = computed(() => url.value && url.value.toLocaleLowerCase().startsWith('http'))

const handleSubmit = async () => {
  const subscribeID = sampleID()

  const subscribe: Subscription = {
    id: subscribeID,
    name: subscribeID,
    useInternal: false,
    url: url.value,
    upload: 0,
    download: 0,
    total: 0,
    expire: 0,
    updateTime: 0,
    type: 'Http',
    website: '',
    path: `data/subscribes/${subscribeID}.yaml`,
    include: '',
    exclude: '',
    includeProtocol: '',
    excludeProtocol: '',
    proxyPrefix: '',
    disabled: false,
    inSecure: false,
    userAgent: '',
    script: DefaultSubscribeScript,
    healthCheck: {
      enable: false,
      url: 'https://www.gstatic.com/generate_204',
      interval: 300,
    },
    proxies: [],
  }

  loading.value = true

  try {
    await subscribeStore.addSubscribe(subscribe)
    await subscribeStore.updateSubscribe(subscribe.id)
  } catch (error: any) {
    loading.value = false
    console.log(error)
    message.error(error)
    subscribeStore.deleteSubscribe(subscribeID)
    return
  }

  const ids = [sampleID(), sampleID(), sampleID(), sampleID(), sampleID()]
  const profile: ProfileType = {
    id: sampleID(),
    name: sampleID(),
    generalConfig: Defaults.GeneralConfigDefaults(),
    advancedConfig: Defaults.AdvancedConfigDefaults(),
    tunConfig: Defaults.TunConfigDefaults(),
    dnsConfig: Defaults.DnsConfigDefaults(),
    proxyGroupsConfig: Defaults.ProxyGroupsConfigDefaults(ids),
    rulesConfig: Defaults.RulesConfigDefaults(ids),
    mixinConfig: Defaults.MixinConfigDefaults(),
    scriptConfig: Defaults.ScriptConfigDefaults(),
  }

  profile.proxyGroupsConfig[0].use = [subscribeID]
  profile.proxyGroupsConfig[1].use = [subscribeID]

  await profilesStore.addProfile(profile)

  appSettingsStore.app.kernel.profile = profile.id

  message.success('home.initSuccessful')

  loading.value = false

  handleCancel()
}
</script>

<template>
  <div class="form-item">
    <div>{{ t('subscribe.url') }} *</div>
    <Input v-model="url" auto-size placeholder="http(s)://" autofocus style="width: 76%" />
  </div>

  <div class="form-action">
    <Button @click="handleCancel" :disabled="loading" type="text">{{ t('common.cancel') }}</Button>
    <Button @click="handleSubmit" :disabled="!canSubmit" :loading="loading" type="primary">
      {{ t('common.save') }}
    </Button>
  </div>
</template>
