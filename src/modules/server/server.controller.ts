import {
  Controller,
  Get,
  Post,
  Body,
  Param,
  Delete,
  Query,
  ParseIntPipe,
  Req
} from '@nestjs/common'
import { ServerService } from './server.service'
import { CreateServerDto, FindPageDto } from './server.dto'
import { ApiBearerAuth } from '@nestjs/swagger'
import { Authorized } from '../../decorators/autrorized.decorator'
import type { User } from '../../../generated/prisma/client'
import { Public } from '../../decorators/public.decorator'
import type { Request } from 'express'

@Controller('server')
export class ServerController {
  constructor(private readonly serverService: ServerService) {}

  @ApiBearerAuth()
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
  findPage(@Authorized() user: User | undefined, @Body() dto: FindPageDto) {
    return this.serverService.findPage(dto, user)
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
