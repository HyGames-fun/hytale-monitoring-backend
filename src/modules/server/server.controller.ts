import {
  Controller,
  Get,
  Post,
  Body,
  Param,
  Delete,
  Query,
  ParseIntPipe,
  Req, ParseArrayPipe,
} from '@nestjs/common'
import { ServerService } from './server.service'
import { CreateServerDto, FindPageDto } from './server.dto'
import { Authorized } from '../../decorators/autrorized.decorator'
import type { User } from '../../../generated/prisma/client'
import { Public } from '../../decorators/public.decorator'
import type { Request } from 'express'
import { Type } from 'class-transformer'

@Controller('server')
export class ServerController {
  constructor(private readonly serverService: ServerService) {}

  @Public()
  @Post('create')
  create(@Body() dto: CreateServerDto) {
    return this.serverService.create(dto)
  }

  @Public()
  @Post('like')
  like(
    @Req() req: Request,
    @Query('id', ParseIntPipe) id: number,
    @Query('token') token?: string,
    @Authorized() user?: User
  ) {
    return this.serverService.like(id, user, token, req)
  }

  @Public()
  @Post()
  findPage(
    @Body() dto: FindPageDto,
    @Req() req: Request,
    @Authorized() user?: User
  ) {
    return this.serverService.findPage(dto, req, user)
  }

  @Public()
  @Get('quantity')
  findQuantity() {
    return this.serverService.findQuantity()
  }

  @Public()
  @Get('statuses')
  getStatuses(
    @Query('ids', new ParseArrayPipe({ items: Number })) ids: number[]
  ) {
    return this.serverService.getOnlineStatuses(ids)
  }

  @Public()
  @Get('statuses')
  getStatus(
    @Query('id', ParseIntPipe) id: number
  ) {
    return this.serverService.getOnlineStatus(id)
  }

  @Public()
  @Get('id/:nameId')
  findOne(
    @Param('nameId') nameId: string,
    @Req() req: Request,
    @Authorized() user?: User
  ) {
    return this.serverService.findOne(nameId, user, req)
  }

  @Delete('id/:id')
  remove(@Param('id') id: string) {
    return this.serverService.remove(+id)
  }
}
