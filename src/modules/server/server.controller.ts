import { Controller, Get, Post, Body, Param, Delete } from '@nestjs/common'
import { ServerService } from './server.service'
import { ServerDto } from './server.dto'

@Controller('server')
export class ServerController {
  constructor(private readonly serverService: ServerService) {}

  @Post()
  create(@Body() createServerDto: ServerDto) {
    return this.serverService.create(createServerDto)
  }

  @Get()
  findAll() {
    return this.serverService.findAll()
  }

  @Get(':id')
  findOne(@Param('id') id: string) {
    return this.serverService.findOne(+id)
  }

  @Delete(':id')
  remove(@Param('id') id: string) {
    return this.serverService.remove(+id)
  }
}
