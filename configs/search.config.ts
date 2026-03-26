import { ConfigService } from '@nestjs/config'
import { ClientOptions, HttpConnection } from '@elastic/elasticsearch'

export function getSearchConfig(configService: ConfigService): ClientOptions {
  return {
    node: configService.getOrThrow('ELASTIC_SEARCH_URL'),
    Connection: HttpConnection
  }
}