import {
  BadRequestException,
  HttpStatus,
  Injectable,
  NotFoundException,
  OnModuleInit
} from '@nestjs/common'
import { ElasticsearchService } from '@nestjs/elasticsearch'
import { ServerDto } from '../server/server.dto'
import { ConfigService } from '@nestjs/config'

@Injectable()
export class SearchService implements OnModuleInit {
  private readonly SEARCH_METHOD: string

  constructor(
    private readonly elasticsearchService: ElasticsearchService,
    private readonly configService: ConfigService
  ) {
    this.SEARCH_METHOD = configService.getOrThrow('SEARCH_METHOD')
  }

  async onModuleInit() {
    if (this.SEARCH_METHOD !== 'elasticsearch') return
    const indexExists = await this.elasticsearchService.indices.exists({
      index: 'servers'
    })

    if (!indexExists) {
      await this.createIndex()
    }
  }

  async createIndex() {
    await this.elasticsearchService.indices.create({
      index: 'servers',
      mappings: {
        properties: {
          name: { type: 'text' },
          domain: { type: 'text' },
          description: { type: 'text' },
          tags: { type: 'text' }
        }
      },
      settings: {
        number_of_shards: 1,
        number_of_replicas: 0
      }
    })
  }

  async indexServer(server: ServerDto) {
    if (server.id == null) {
      throw new Error('Server id is required for Elasticsearch indexing')
    }

    return this.elasticsearchService.index({
      index: 'servers',
      id: server.id.toString(),
      document: {
        name: server.name,
        description: server.description,
        domain: server.domain,
        tags: server.tags
      }
    })
  }

  async deleteServer(id: number) {
    try {
      await this.elasticsearchService.delete({
        index: 'servers',
        id: id.toString()
      })
    } catch (e) {
      if (e.meta?.statusCode === HttpStatus.NOT_FOUND) {
        throw new NotFoundException('Server not found')
      }
      throw e
    }
  }

  async searchServers(query: string) {
    if (!query) throw new BadRequestException('Query is empty')

    const result = await this.elasticsearchService.search({
      index: 'servers',
      query: {
        bool: {
          should: [
            {
              multi_match: {
                query,
                fields: ['name^3', 'description', 'domain', 'tags^2'],
                fuzziness: 'AUTO'
              }
            },
            {
              prefix: {
                name: query.toLowerCase()
              }
            }
          ]
        }
      }
    })

    return result.hits.hits.map((hit) => hit._source)
  }
}
