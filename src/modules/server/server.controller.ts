import { Controller, Get, Post, Body, Param, Delete, Query, ParseIntPipe } from '@nestjs/common'
import { ServerService } from './server.service'
import { CreateServerDto } from './server.dto'
import { ApiBearerAuth } from '@nestjs/swagger'
import { Authorized } from '../../decorators/autrorized.decorator'
import type { User } from '../../../generated/prisma/client'

@Controller('server')
export class ServerController {
  constructor(private readonly serverService: ServerService) {}

  @ApiBearerAuth()
  @Post('create')
  create(@Body() dto: CreateServerDto) {
    return this.serverService.create(dto)
  }

  @ApiBearerAuth()
  @Post('like')
  like(
    @Query('id', ParseIntPipe) id: number,
    @Authorized() user: User
  ) {
    return this.serverService.like(id, user)
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
