import { ConfigService } from '@nestjs/config'
import { HttpModuleOptions } from '@nestjs/axios'

export function getHttpConfig(configService: ConfigService): HttpModuleOptions {
  return {
    baseURL: configService.getOrThrow('DISCORD_OAUTH_API_URL')
  }
}