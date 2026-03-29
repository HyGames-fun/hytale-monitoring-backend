import {
  Body,
  Controller,
  Get,
  HttpStatus,
  ParseIntPipe,
  Post,
  Query,
  Redirect,
  Req,
  Res
} from '@nestjs/common'
import { AuthService } from './auth.service'
import type { Response, Request } from 'express'
import { LoginDto, RegisterDto } from './auth.dto'
import { UndefinedPipe } from '../../pipes/undefined.pipe'
import { Public } from '../../decorators/public.decorator'

@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @Public()
  @Post('register')
  register(
    @Res({ passthrough: true }) res: Response,
    @Req() req: Request,
    @Query('code', ParseIntPipe) code: number
  ) {
    return this.authService.register(res, req, code)
  }

  @Public()
  @Post('init-register')
  async initRegister(
    @Res({ passthrough: true }) res: Response,
    @Body() dto: RegisterDto
  ) {
    await this.authService.sendVerificationEmail(dto, res)
  }

  @Public()
  @Post('resend-email')
  async resendEmail(@Req() req: Request) {
    await this.authService.resendVerificationEmail(req)
  }

  @Public()
  @Post('login')
  login(@Res({ passthrough: true }) res: Response, @Body() dto: LoginDto) {
    return this.authService.login(res, dto)
  }

  @Public()
  @Post('logout')
  logout(@Res({ passthrough: true }) res: Response) {
    this.authService.logout(res)
  }

  @Public()
  @Post('refresh')
  refresh(@Res({ passthrough: true }) res: Response, @Req() req: Request) {
    return this.authService.refresh(res, req)
  }

  @Public()
  @Get('discord/login')
  @Redirect(
    'https://discord.com/oauth2/authorize?client_id=1475225594282381513&response_type=code&redirect_uri=http%3A%2F%2Flocalhost%3A4242%2Fapi%2Fauth%2Fdiscord%2Fcallback&scope=identify',
    HttpStatus.FOUND
  )
  discordLogin() {}

  @Public()
  @Get('discord/callback')
  discordCallback(
    @Res({ passthrough: true }) res: Response,
    @Query('code', UndefinedPipe) code: string
  ) {
    return this.authService.discordCallback(res, code)
  }
}
