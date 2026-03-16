import {
  Controller,
  Get,
  Post,
  Body,
  Param,
  Delete,
  Query,
  ParseIntPipe
} from '@nestjs/common'
import { ServerService } from './server.service'
import { CreateServerDto, FindPageDto } from './server.dto'
import { ApiBearerAuth } from '@nestjs/swagger'
import { Authorized } from '../../decorators/autrorized.decorator'
import type { User } from '../../../generated/prisma/client'
import { Public } from '../../decorators/public.decorator'

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
  like(@Query('id', ParseIntPipe) id: number, @Authorized() user: User) {
    return this.serverService.like(id, user)
  }

  @Public()
  @Post()
  findPage(@Body() dto: FindPageDto) {
    return this.serverService.findPage(dto)
  }

  @Public()
  @Get('quantity')
  findQuantity() {
    return this.serverService.findQuantity()
  }

  @Get('id/:id')
  findOne(@Param('id') id: string) {
    return this.serverService.findOne(+id)
  }

  @Delete('id/:id')
  remove(@Param('id') id: string) {
    return this.serverService.remove(+id)
  }
}
