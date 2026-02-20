import { ConfigService } from '@nestjs/config';
import { JwtModuleOptions } from '@nestjs/jwt';

export async function getJwtConfig(configService: ConfigService): Promise<JwtModuleOptions> {
  return {
    secret: await configService.getOrThrow('JWT_SECRET'),
    signOptions: {
      algorithm: 'HS256'
    },
    verifyOptions: {
      algorithms: ['HS256'],
      ignoreExpiration: false
    }
  }
}