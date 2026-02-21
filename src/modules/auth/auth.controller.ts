import { Body, Controller, Post, Req, Res } from '@nestjs/common'
import { AuthService } from './auth.service'
import type { Response, Request } from 'express'
import { LoginDto, RegisterDto } from './auth.dto'

@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @Post('register')
  register(
    @Res({ passthrough: true }) res: Response,
    @Body() dto: RegisterDto,
  ) {
    return this.authService.register(res, dto)
  }

  @Post('login')
  login(@Res({ passthrough: true }) res: Response, @Body() dto: LoginDto) {
    return this.authService.login(res, dto)
  }

  @Post('logout')
  logout(@Res({ passthrough: true }) res: Response) {
    this.authService.logout(res)
  }

  @Post('refresh')
  refresh(@Res({ passthrough: true }) res: Response, @Req() req: Request) {
    return this.authService.refresh(res, req)
  }
}
