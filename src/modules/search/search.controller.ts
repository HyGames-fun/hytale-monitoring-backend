import { Controller, Get, ParseIntPipe, Query } from '@nestjs/common'
import { SearchService } from './search.service'
import { Public } from '../../decorators/public.decorator'

@Controller('search')
export class SearchController {
  constructor(private readonly searchService: SearchService) {}

  @Public()
  @Get()
  search(
    @Query('query') q: string,
    @Query('page', ParseIntPipe) page: number,
    @Query('limit', ParseIntPipe) limit: number
  ) {
    return this.searchService.searchServers(q, page, limit)
  }
}
