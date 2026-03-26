import { Module } from '@nestjs/common'
import { SearchService } from './search.service'
import { SearchController } from './search.controller'
import { ElasticsearchModule } from '@nestjs/elasticsearch'
import { ConfigService } from '@nestjs/config'
import { getSearchConfig } from '../../../configs/search.config'

@Module({
  imports: [
    ElasticsearchModule.registerAsync({
      useFactory: getSearchConfig,
      inject: [ConfigService]
    })
  ],
  controllers: [SearchController],
  providers: [SearchService],
  exports: [SearchService]
})
export class SearchModule {}
