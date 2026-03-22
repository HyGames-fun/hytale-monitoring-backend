import {
  Body,
  Controller,
  Delete,
  FileTypeValidator,
  Get,
  MaxFileSizeValidator,
  Param,
  ParseFilePipe,
  ParseIntPipe,
  Put,
  UploadedFile,
  UseInterceptors
} from '@nestjs/common'
import { UserService } from './user.service'
import { Authorized } from '../../decorators/autrorized.decorator'
import type { User } from '../../../generated/prisma/client'
import { ApiBearerAuth } from '@nestjs/swagger'
import { FileInterceptor } from '@nestjs/platform-express'
import { RegionDto } from './user.dto'
import { RegionPipe } from './pipes/region.pipe'

@Controller('user')
export class UserController {
  constructor(private readonly userService: UserService) {}

  @ApiBearerAuth()
  @Get()
  getCurrentUser(@Authorized() user: User) {
    return this.userService.getUser(user)
  }

  @Put('avatar')
  @UseInterceptors(FileInterceptor('file'))
  async setAvatar(
    @UploadedFile(
      new ParseFilePipe({
        fileIsRequired: true,
        validators: [
          new MaxFileSizeValidator({ maxSize: 3000000 }),
          new FileTypeValidator({
            fileType: /(image\/jpeg|image\/png|image\/webp)$/
          })
        ]
      })
    )
    file: Express.Multer.File,
    @Body('region', new RegionPipe()) region: any,
    @Authorized() user: User
  ) {
    await this.userService.setAvatar(file.buffer, user.id, region as RegionDto)
  }

  @Delete('avatar')
  async deleteAvatar(@Authorized() user: User) {
    await this.userService.deleteAvatar(user.id, user.avatar)
  }

  @Get('avatar')
  async getSelfAvatar(@Authorized() user: User) {
    await this.userService.getAvatar(user.id)
  }

  @Get('avatar/:id')
  async getAvatar(@Param('id', ParseIntPipe) userId: number) {
    await this.userService.getAvatar(userId)
  }
}
